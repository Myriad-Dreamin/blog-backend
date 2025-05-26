package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/mail"
	"net/url"
)

type UnlikeHandler struct {
	h      *Handler
	isPost bool
}

func (h UnlikeHandler) handleLike(w http.ResponseWriter, r *http.Request) {
	h.h.handleLike(w, r, h.isPost)
}

func (h *Handler) handleLike(w http.ResponseWriter, r *http.Request, isPost bool) {
	// post or get
	switch r.Method {
	case http.MethodGet:
		h.handleLikeGet(w, r)
	case http.MethodPost:
		// rate limit
		remaining := h.reactionLim.Reserve()
		if !remaining.OK() {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		h.handleLikePost(w, r, isPost)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleLikeGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get article ID from query parameters
	articleId := r.URL.Query().Get("articleId")
	if articleId == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	// decode url component
	email, err := url.QueryUnescape(email)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	mailAddr, err := mail.ParseAddress(email)
	if err != nil {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	var exists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM likes WHERE article_id=? AND email=?)", articleId, mailAddr.Address).Scan(&exists)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Return like status as JSON
	type Status struct {
		Exists bool `json:"exists"`
	}

	status := Status{
		Exists: exists,
	}

	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) handleLikePost(w http.ResponseWriter, r *http.Request, isPost bool) {
	// Get article ID from form data
	articleId := r.FormValue("articleId")
	if articleId == "" {
		http.Error(w, "Empty articleId", http.StatusBadRequest)
		return
	}
	// Check if article exists in database
	if !h.mustExistsArticle(articleId, w) {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	// Email
	email := r.FormValue("email")
	if email == "" {
		http.Error(w, "Empty email", http.StatusBadRequest)
		return
	}

	// Validate article ID
	if len(articleId) > 64 {
		http.Error(w, "Article ID too long", http.StatusBadRequest)
		return
	}

	// Validate email format
	if len(email) > 128 {
		http.Error(w, "Email too long", http.StatusBadRequest)
		return
	}

	// Validate email format (RFC 5322 address)
	mailAddr, err := mail.ParseAddress(email)
	if err != nil {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}
	email = mailAddr.Address

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		log.Printf("error begin transaction: %s\n", err)
	}
	defer func() {
		// tx.Rollback()
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("error rolling back transaction: %s\n", err)
		}
	}()

	if isPost {
		var exists bool
		err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM likes WHERE article_id=? AND email=?)", articleId, email).Scan(&exists)
		if err != nil {
			log.Printf("error checking like existence: %s\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if exists {
			http.Error(w, "Already liked", http.StatusBadRequest)
			return
		}

		// Insert Like into database
		_, err = tx.Exec("INSERT INTO likes (article_id, email) VALUES (?, ?)", articleId, email)
		if err != nil {
			log.Printf("error inserting like: %s\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))

		log.Printf("Like added to article: %v, %v\n", articleId, email)
	} else {
		// Delete Like from database
		_, err = tx.Exec("DELETE FROM likes WHERE article_id=? AND email=?", articleId, email)
		if err != nil {
			log.Printf("error deleting like: %s\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))

		log.Printf("Like removed from article: %v, %v\n", articleId, email)
	}

	if err = tx.Commit(); err != nil {
		log.Printf("error committing transaction: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
