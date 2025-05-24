package main

import (
	"database/sql"
	"log"
	"net/mail"
	"time"

	"github.com/Myriad-Dreamin/blog-backend/pkg/dto"
	"github.com/Myriad-Dreamin/blog-backend/pkg/iou"
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

func (h *Handler) tickSnapshot() {
	// periodly snapshot
	// every day
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	log.Printf("Snapshot...\n")
	h.writeSnapshot()

	for range ticker.C {
		log.Printf("Snapshot...\n")
		h.writeSnapshot()
	}
}

func (h *Handler) writeSnapshot() {
	// export clicks to `.data/article-clicks.json`
	{
		clicks, err := sqlite.GetClicks(h.db)
		if err != nil {
			log.Printf("error get clicks: %s\n", err)
		}
		log.Printf("write clicks: %v\n", len(clicks))

		// write clicks to file
		err = iou.WriteJsonToFile("./.data/article-clicks.json", clicks)
		if err != nil {
			log.Printf("error write build info to file: %s\n", err)
		}
	}
	// export comments to `.data/article-comments.json`
	{
		comments, err := sqlite.GetComments(h.db)
		if err != nil {
			log.Printf("error get clicks: %s\n", err)

		}
		log.Printf("write comments: %v\n", len(comments))

		// write clicks to file
		err = iou.WriteJsonToFile("./.data/article-email-comments.json", comments)
		if err != nil {
			log.Printf("error write build info to file: %s\n", err)
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

		// write clicks to file
		err = iou.WriteJsonToFile("./.data/article-comments.json", comments)
		if err != nil {
			log.Printf("error write build info to file: %s\n", err)
		}
	}

	// 	@cp .data/article-clicks.json $(BLOG_FRONTEND_PATH)/content/snapshot/article-clicks.json
	// 	@cp .data/article-comments.json $(BLOG_FRONTEND_PATH)/content/snapshot/article-comments.json
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
	h.articles = articles
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
