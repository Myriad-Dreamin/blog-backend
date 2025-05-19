package main

import (
	"database/sql"
	"net/mail"

	"github.com/Myriad-Dreamin/blog-backend/pkg/iou"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var cmdComment = &cobra.Command{
	Use:   "comment",
	Short: "extract article comments",
	RunE: func(cmd *cobra.Command, args []string) error {
		// open db
		db, err := sql.Open("sqlite3", "./.data/blog.db")
		if err != nil {
			return errors.Wrap(err, "error open db")

		}
		defer db.Close()

		// export comments to `.data/article-comments.json`
		comments, err := getComments(db)
		if err != nil {
			return errors.Wrap(err, "error get clicks")

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
			return errors.Wrap(err, "error write build info to file")
		}

		return nil
	},
}

func getComments(db *sql.DB) ([]ArticleComment, error) {
	rows, err := db.Query("SELECT id, article_id, content, email, created_at FROM comments")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var comments []ArticleComment
	for rows.Next() {
		var comment ArticleComment
		if err := rows.Scan(&comment.Id, &comment.ArticleId, &comment.Content, &comment.Email, &comment.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return comments, nil
}

type ArticleComment struct {
	Id        string `json:"id"`
	ArticleId string `json:"articleId"`
	Content   string `json:"content"`
	Email     string `json:"email"`
	CreatedAt int64  `json:"createdAt"`
}
