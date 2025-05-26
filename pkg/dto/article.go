package dto

import (
	"encoding/json"
	"os"
)

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

type Article struct {
	Id string `json:"id"`
}

type ArticleStat struct {
	Id    string `json:"id"`
	Click int    `json:"click"`
	Like  int    `json:"like"`
}

type ArticleComment struct {
	Id        string `json:"id"`
	ArticleId string `json:"articleId"`
	Content   string `json:"content"`
	Email     string `json:"email"`
	CreatedAt int64  `json:"createdAt"`
}
