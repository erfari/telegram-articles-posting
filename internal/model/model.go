package model

import "time"

type Item struct {
	Title      string
	Categories []string
	Link       string
	Date       time.Time
	Summary    string
	SourceName string
}

type Source struct {
	ID        int64 //uuid добавыить
	Name      string
	FeedURL   string
	CreatedAt time.Time
}

type Article struct {
	ID          int64 //uuid add
	SourceID    int64
	Title       string
	Link        string
	Summary     string
	PublishedAt time.Time
	PostedAt    time.Time
	CreateAt    time.Time
}
