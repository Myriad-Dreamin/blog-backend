package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/Myriad-Dreamin/blog-backend/pkg/dto"
	"github.com/Myriad-Dreamin/blog-backend/pkg/iou"
	"github.com/pkg/errors"
)

func BackupBlog(src *sql.DB) error {
	info, available := debug.ReadBuildInfo()

	if !available {
		log.Println("Build info not available")
	}

	// make directory
	err := os.MkdirAll("./.data/backup/tmp", 0755)
	if err != nil {
		return errors.Wrap(err, "error create directory")
	}

	// write build info to file
	err = iou.WriteJsonToFile("./.data/backup/tmp/build-info.json", info)
	if err != nil {
		return errors.Wrap(err, "error write build info to file")
	}

	if src == nil {
		db, err := sql.Open("sqlite3", "./.data/blog.db")
		if err != nil {
			return errors.Wrap(err, "error open db")
		}
		defer db.Close()
		src = db
	}

	// open target db
	destDb, err := sql.Open("sqlite3", "./.data/backup/tmp/blog.db")
	if err != nil {
		return errors.Wrap(err, "error open target db")

	}
	defer destDb.Close()

	// backup db
	err = Backup(destDb, src)
	if err != nil {
		return errors.Wrap(err, "error backup db")
	}

	// close target db
	err = destDb.Close()
	if err != nil {
		return errors.Wrap(err, "error close target db")
	}

	// move target db to backup
	var timestamp = time.Now().Format("2006-01-02_15-04-05")
	err = os.Rename("./.data/backup/tmp", fmt.Sprintf("./.data/backup/%s", timestamp))
	if err != nil {
		return errors.Wrap(err, "error move target db to backup")
	}

	return nil
}

func GetStats(db *sql.DB) ([]dto.ArticleStat, error) {
	rows, err := db.Query("SELECT id, click, COALESCE(like, 0) FROM articles LEFT JOIN (SELECT article_id, COUNT(*) AS like FROM likes GROUP BY article_id) AS like_stats ON articles.id = like_stats.article_id")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var clicks []dto.ArticleStat
	for rows.Next() {
		var click dto.ArticleStat
		if err := rows.Scan(&click.Id, &click.Click, &click.Like); err != nil {
			return nil, err
		}
		clicks = append(clicks, click)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return clicks, nil
}

func GetComments(db *sql.DB) ([]dto.ArticleComment, error) {
	rows, err := db.Query("SELECT id, article_id, content, email, authorized, created_at FROM comments")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var comments []dto.ArticleComment
	for rows.Next() {
		var comment dto.ArticleComment
		if err := rows.Scan(&comment.Id, &comment.ArticleId, &comment.Content, &comment.Email, &comment.Authorized, &comment.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return comments, nil
}
