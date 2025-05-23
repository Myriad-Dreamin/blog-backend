package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Myriad-Dreamin/blog-backend/pkg/dto"
)

func (h *Handler) handleClick(w http.ResponseWriter, r *http.Request) {
	// post or get
	switch r.Method {
	case http.MethodPost:
		h.handleClickPost(w, r)
	case http.MethodGet:
		h.handleClickGet(w, r)
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

	type Status struct {
		Count int `json:"count"`
	}
	status := Status{
		Count: count,
	}

	// Return click count as JSON
	w.Header().Set("Content-Type", "application/json")
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

	var remoteAddr = r.RemoteAddr

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
	log.Printf("Article clicked: %s, from %s\n", article.Id, remoteAddr)
}
