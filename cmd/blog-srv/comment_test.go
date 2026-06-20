package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestHandleCommentPostRejectsOversizedForm(t *testing.T) {
	h := &Handler{}
	form := url.Values{
		"articleId": {"post-1"},
		"content":   {strings.Repeat("a", int(maxCommentFormBytes))},
		"email":     {"alice@example.com"},
	}
	req := httptest.NewRequest(http.MethodPost, "/article/comment", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	h.handleCommentPost(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d; body = %q", rec.Code, http.StatusRequestEntityTooLarge, rec.Body.String())
	}
}

func TestHandleCommentPostCreatesPendingComment(t *testing.T) {
	db := newTestBlogDB(t)
	h := &Handler{db: db}
	form := url.Values{
		"articleId": {"post-1"},
		"content":   {"looks good"},
		"email":     {"Alice <alice@example.com>"},
	}
	req := httptest.NewRequest(http.MethodPost, "/article/comment", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	h.handleCommentPost(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", rec.Code, http.StatusOK, rec.Body.String())
	}

	var content, email string
	var authorized, rejected bool
	err := db.QueryRow("SELECT content, email, authorized, rejected FROM comments WHERE article_id = ?", "post-1").
		Scan(&content, &email, &authorized, &rejected)
	if err != nil {
		t.Fatalf("query comment: %v", err)
	}
	if content != "looks good" {
		t.Fatalf("content = %q, want %q", content, "looks good")
	}
	if email != "Alice <alice@example.com>" {
		t.Fatalf("email = %q, want %q", email, "Alice <alice@example.com>")
	}
	if authorized || rejected {
		t.Fatalf("authorized/rejected = %v/%v, want false/false", authorized, rejected)
	}
}

func TestHandleCommentPostRejectsInvalidEmailDisplayName(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{name: "missing display name", email: "alice@example.com"},
		{name: "empty display name", email: `"" <alice@example.com>`},
		{name: "non-printable display name", email: "\"Alice\x7f\" <alice@example.com>"},
		{name: "reserved display name character", email: `"Alice [admin]" <alice@example.com>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestBlogDB(t)
			h := &Handler{db: db}
			form := url.Values{
				"articleId": {"post-1"},
				"content":   {"looks good"},
				"email":     {tt.email},
			}
			req := httptest.NewRequest(http.MethodPost, "/article/comment", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()

			h.handleCommentPost(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body = %q", rec.Code, http.StatusBadRequest, rec.Body.String())
			}

			var count int
			if err := db.QueryRow("SELECT COUNT(*) FROM comments").Scan(&count); err != nil {
				t.Fatalf("count comments: %v", err)
			}
			if count != 0 {
				t.Fatalf("comments count = %d, want 0", count)
			}
		})
	}
}

func newTestBlogDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close db: %v", err)
		}
	})

	stmts := []string{
		"CREATE TABLE articles (id TEXT PRIMARY KEY, click INTEGER DEFAULT 0)",
		"CREATE TABLE comments (id INTEGER PRIMARY KEY AUTOINCREMENT, article_id TEXT, email TEXT, content TEXT, authorized BOOLEAN NOT NULL DEFAULT FALSE, rejected BOOLEAN NOT NULL DEFAULT FALSE, created_at INTEGER)",
		"INSERT INTO articles (id) VALUES ('post-1')",
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("exec %q: %v", stmt, err)
		}
	}

	return db
}
