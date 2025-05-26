package main

import (
	"net/http"

	"github.com/Myriad-Dreamin/blog-backend/pkg/sqlite"
)

func (h *Handler) handleSnapshotStats(w http.ResponseWriter, r *http.Request) {
	if h.rateLimit(w) {
		h.jsonGet(w, r, func() (any, error) {
			return sqlite.GetStats(h.db)
		})
	}
}

func (h *Handler) handleSnapshotComments(w http.ResponseWriter, r *http.Request) {
	if h.rateLimit(w) {
		h.jsonGet(w, r, func() (any, error) {
			return sqlite.GetComments(h.db)
		})
	}
}
