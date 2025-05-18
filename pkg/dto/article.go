package dto

import (
	"encoding/json"
	"os"
)

type Article struct {
	Id string `json:"id"`
}

func GetArticles() ([]Article, error) {
	var articles []Article

	file, err := os.Open("./.data/articles.json")
	if err != nil {
		return nil, err
	}

	defer file.Close()
	err = json.NewDecoder(file).Decode(&articles)
	if err != nil {
		return nil, err
	}
	return articles, nil
}
