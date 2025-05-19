package main

import (
	"database/sql"

	"github.com/Myriad-Dreamin/blog-backend/pkg/iou"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var cmdClick = &cobra.Command{
	Use:   "click",
	Short: "extract article clicks",
	RunE: func(cmd *cobra.Command, args []string) error {
		// open db
		db, err := sql.Open("sqlite3", "./.data/blog.db")
		if err != nil {
			return errors.Wrap(err, "error open db")

		}
		defer db.Close()

		// export clicks to `.data/article-clicks.json`
		clicks, err := getClicks(db)
		if err != nil {
			return errors.Wrap(err, "error get clicks")

		}

		// write clicks to file
		err = iou.WriteJsonToFile("./.data/article-clicks.json", clicks)
		if err != nil {
			return errors.Wrap(err, "error write build info to file")
		}

		return nil
	},
}

func getClicks(db *sql.DB) ([]ArticleClick, error) {
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
