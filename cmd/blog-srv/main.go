package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"database/sql"

	"github.com/Myriad-Dreamin/blog-backend/pkg/dto"
	"github.com/rs/cors"

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
		db: db,
	}

	h.createArticles()

	mux := http.NewServeMux()

	mux.HandleFunc("/article/click", h.handleClick)

	handler := cors.Default().Handler(mux)
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
}

func (h *Handler) createArticles() {
	// .data/articles.json
	articles, err := dto.GetArticles()
	if err != nil {
		log.Printf("error get articles: %s\n", err)
		return
	}

	// Create table if not exists
	_, err = h.db.Exec("CREATE TABLE IF NOT EXISTS articles (id TEXT PRIMARY KEY, click INTEGER DEFAULT 0)")
	if err != nil {
		log.Printf("error creating table: %s\n", err)
		return
	}
	// Insert articles into database
	for _, article := range articles {
		_, err = h.db.Exec("INSERT OR IGNORE INTO articles (id) VALUES (?)", article.Id)
		if err != nil {
			log.Printf("error inserting article: %s\n", err)
			return
		}
	}

	log.Printf("Articles loaded: %d\n", len(articles))
}

func (h *Handler) handleClick(w http.ResponseWriter, r *http.Request) {
	// post or get
	switch r.Method {
	case http.MethodPost:
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		h.handleClickPost(w, r)
	case http.MethodGet:
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		h.handleClickGet(w, r)
	case http.MethodOptions:
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleClickGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get article ID from query parameters
	articleId := r.URL.Query().Get("id")
	if articleId == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Gets click count from database
	var count int
	err := h.db.QueryRow("SELECT click FROM articles WHERE id=?", articleId).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	// Return click count as JSON
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	type Status struct {
		Count int `json:"count"`
	}
	// Return click count as JSON
	status := Status{
		Count: count,
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) handleClickPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var article dto.Article
	err := json.NewDecoder(r.Body).Decode(&article)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if article.Id == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	// Check if article exists in database
	var exists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM articles WHERE id=?)", article.Id).Scan(&exists)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	// Increment click count
	_, err = h.db.Exec("UPDATE articles SET click = click + 1 WHERE id=?", article.Id)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	type Status struct {
		Message string `json:"message"`
	}

	status := Status{
		Message: "Successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log the click
	log.Printf("Article clicked: %s\n", article.Id)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
