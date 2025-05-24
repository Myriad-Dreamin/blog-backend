package main

import (
	"log"
	"net/http"
	"net/mail"
	"time"
)

func (h *Handler) handleComment(w http.ResponseWriter, r *http.Request) {
	// post or get
	switch r.Method {
	case http.MethodPost:
		// rate limit
		remaining := h.commentLim.Reserve()
		if !remaining.OK() {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		h.handleCommentPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleCommentPost(w http.ResponseWriter, r *http.Request) {
	createdAt := time.Now().UnixMilli()
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

	// Get comment from form data
	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Empty content", http.StatusBadRequest)
		return
	}

	// Email
	email := r.FormValue("email")
	if email == "" {
		http.Error(w, "Empty email", http.StatusBadRequest)
		return
	}

	// Validate comment length
	if len(content) > 4096 {
		http.Error(w, "Comment too long", http.StatusBadRequest)
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
	_, err := mail.ParseAddress(email)
	if err != nil {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// Insert comment into database
	_, err = h.db.Exec("INSERT INTO comments (article_id, content, email, created_at) VALUES (?, ?, ?, ?)", articleId, content, email, createdAt)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))

	log.Printf("Comment added to article: %v\n", articleId)
}
