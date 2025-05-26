package main

import (
	"net/http"
	"net/mail"

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
			comments, err := sqlite.GetComments(h.db)
			if comments == nil {
				return nil, err
			}

			var withEmail = false
			if !withEmail {
				for i := range comments {
					addr, err := mail.ParseAddress(comments[i].Email)
					if err != nil {
						comments[i].Email = ""
					}
					comments[i].Email = addr.Name
				}
			}

			return comments, nil
		})
	}
}
