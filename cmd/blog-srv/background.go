package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/Myriad-Dreamin/blog-backend/pkg/dto"
	"github.com/Myriad-Dreamin/blog-backend/pkg/sqlite"
	"github.com/fsnotify/fsnotify"
)

func (h *Handler) watchArticles() {
	h.createTables()
	// watch articles
	notifier, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("error creating watcher: %s\n", err)
		return
	}
	defer notifier.Close()

	err = notifier.Add("./.data/articles.json")
	if err != nil {
		log.Printf("error adding watcher: %s\n", err)
		return
	}
	for {
		select {
		case event, ok := <-notifier.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Printf("File modified: %s\n", event.Name)
				h.createTables()
			}
		case err, ok := <-notifier.Errors:
			if !ok {
				return
			}
			log.Printf("error: %s\n", err)
		}
	}
}

func (h *Handler) createTables() {
	articles, err := dto.GetArticles()
	if err != nil {
		log.Printf("error get articles: %s\n", err)
		return
	}

	h.createArticles(articles)
	h.createComments()
}

func (h *Handler) createArticles(articles []dto.Article) {
	// Create table if not exists
	_, err := h.db.Exec("CREATE TABLE IF NOT EXISTS articles (id TEXT PRIMARY KEY, click INTEGER DEFAULT 0)")
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

func (h *Handler) createComments() {
	// Create table if not exists
	_, err := h.db.Exec("CREATE TABLE IF NOT EXISTS comments (id INTEGER PRIMARY KEY AUTOINCREMENT, article_id TEXT, email TEXT, content TEXT, created_at INTEGER)")
	if err != nil {
		log.Printf("error creating table: %s\n", err)
		return
	}
	// Create index if not exists
	_, err = h.db.Exec("CREATE INDEX IF NOT EXISTS idx_article_id ON comments (article_id)")
	if err != nil {
		log.Printf("error creating index: %s\n", err)
		return
	}
}

func backup(conn *sql.DB) {
	// periodly backup
	// every day
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		// backup
		log.Printf("Backing up...\n")
		sqlite.BackupBlog(conn)
	}
}
