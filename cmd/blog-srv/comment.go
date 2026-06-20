package main

import (
	"errors"
	"log"
	"mime"
	"net/http"
	"time"

	"github.com/Myriad-Dreamin/blog-backend/pkg/commentmoderation"
)

const (
	maxCommentContentBytes   = 4096
	maxCommentArticleIDBytes = 64
	maxCommentEmailBytes     = 128

	// URL-encoded form bodies can expand each validated byte to %XX.
	maxCommentFormBytes int64 = 16 << 10
)

func (h *Handler) handleComment(w http.ResponseWriter, r *http.Request) {
	// post or get
	switch r.Method {
	case http.MethodPost:
		if h.rateLimit(w) {
			h.handleCommentPost(w, r)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleCommentPost(w http.ResponseWriter, r *http.Request) {
	if !parseCommentForm(w, r) {
		return
	}

	createdAt := time.Now().UnixMilli()
	// Get article ID from form data
	articleId := r.Form.Get("articleId")
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
	content := r.Form.Get("content")
	if content == "" {
		http.Error(w, "Empty content", http.StatusBadRequest)
		return
	}

	// Email
	email := r.Form.Get("email")
	if email == "" {
		http.Error(w, "Empty email", http.StatusBadRequest)
		return
	}

	// Validate comment length
	if len(content) > maxCommentContentBytes {
		http.Error(w, "Comment too long", http.StatusBadRequest)
		return
	}

	// Validate article ID
	if len(articleId) > maxCommentArticleIDBytes {
		http.Error(w, "Article ID too long", http.StatusBadRequest)
		return
	}

	// Validate email format
	if len(email) > maxCommentEmailBytes {
		http.Error(w, "Email too long", http.StatusBadRequest)
		return
	}

	// Validate email format and the public display name.
	email, err := commentmoderation.ValidateCommentEmail(email)
	if err != nil {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// Insert comment into database
	_, err = h.db.Exec("INSERT INTO comments (article_id, content, email, authorized, rejected, created_at) VALUES (?, ?, ?, ?, ?, ?)", articleId, content, email, false, false, createdAt)
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

func parseCommentForm(w http.ResponseWriter, r *http.Request) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxCommentFormBytes)

	var err error
	mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if mediaType == "multipart/form-data" {
		err = r.ParseMultipartForm(maxCommentFormBytes)
	} else {
		err = r.ParseForm()
	}
	if err == nil {
		return true
	}

	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		return false
	}

	http.Error(w, "Bad request", http.StatusBadRequest)
	return false
}
