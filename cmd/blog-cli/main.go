package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	// open db
	db, err := sql.Open("sqlite3", "./.data/blog.db")
	if err != nil {
		log.Printf("error open db: %s\n", err)
		return
	}

	// export clicks to `.data/article-clicks.json`
	clicks, err := getClicks(db)
	if err != nil {
		log.Printf("error get clicks: %s\n", err)
		return
	}

	// write clicks to file
	file, err := os.Create("./.data/article-clicks.json")
	if err != nil {
		log.Printf("error create file: %s\n", err)
		return
	}
	defer file.Close()
	err = json.NewEncoder(file).Encode(clicks)
	if err != nil {
		log.Printf("error encode clicks: %s\n", err)
		return
	}
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
