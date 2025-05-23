package dto

import (
	"database/sql"
)

func GetComments(db *sql.DB) ([]ArticleComment, error) {
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
