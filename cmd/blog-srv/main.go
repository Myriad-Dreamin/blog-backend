package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"database/sql"

	"github.com/rs/cors"
	"golang.org/x/time/rate"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Printf("Starting blog server...\n")

	// Makes directory
	if _, err := os.Stat("./.data"); os.IsNotExist(err) {
		err := os.Mkdir("./.data", 0755)
		if err != nil {
			log.Printf("error creating directory: %s\n", err)
			os.Exit(1)
		}
	}

	db, err := sql.Open("sqlite3", "./.data/blog.db")
	checkErr(err)
	defer db.Close()

	var h = &Handler{
		db:          db,
		reactionLim: rate.NewLimiter(rate.Every(1), 1),
	}

	go h.watchArticles()
	go h.tickSnapshot()
	go backup(db)

	mux := http.NewServeMux()

	mux.HandleFunc("/article/click", h.handleClick)
	mux.HandleFunc("/article/comment", h.handleComment)
	mux.HandleFunc("/article/like", UnlikeHandler{h, true}.handleLike)
	mux.HandleFunc("/article/like/delete", UnlikeHandler{h, false}.handleLike)
	mux.HandleFunc("/snapshot/stats", h.handleSnapshotStats)
	mux.HandleFunc("/snapshot/comments", h.handleSnapshotComments)

	corsCfg := cors.New(cors.Options{
		AllowedHeaders: []string{"accept", "content-type", "x-requested-with", "referrer-policy"},
	})
	// corsCfg.Log = log.Default()

	handler := corsCfg.Handler(mux)
	err = http.ListenAndServe(":13333", handler)
	if errors.Is(err, http.ErrServerClosed) {
		log.Printf("server closed\n")
	} else if err != nil {
		log.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

type Handler struct {
	db *sql.DB

	reactionLim *rate.Limiter
}

// Checks if article exists in database
func (h *Handler) mustExistsArticle(id string, w http.ResponseWriter) bool {
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM articles WHERE id=?)", id).Scan(&exists)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return false
	}
	if !exists {
		http.Error(w, "Not found", http.StatusNotFound)
		return false
	}

	return true
}

func (h *Handler) jsonGet(w http.ResponseWriter, r *http.Request, getter func() (any, error)) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := getter()
	if err != nil {
		log.Printf("error get stats: %s\n", err)
	}

	// Return click count as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
