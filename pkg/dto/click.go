package dto

import (
	"database/sql"
)

func GetClicks(db *sql.DB) ([]ArticleClick, error) {
	rows, err := db.Query("SELECT id, click FROM articles")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var clicks []ArticleClick
	for rows.Next() {
		var click ArticleClick
		if err := rows.Scan(&click.Id, &click.Click); err != nil {
			return nil, err
		}
		clicks = append(clicks, click)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return clicks, nil
}

type ArticleClick struct {
	Id    string `json:"id"`
	Click int    `json:"click"`
}
