package main

import (
	"time"
)

// Post contains all data for one blog post
type Post struct {
	ID       uint32    `json:"_id"`
	IsShort  bool      `json:"isshort"`
	Title    string    `json:"title"`
	URLTitle string    `json:"urltitle"`
	Visible  bool      `json:"visible"`
	Date     time.Time `json:"date"`
	Body     string    `json:"body"`
	Markdown string    `json:"markdown"`
}

// Posts is just an array of posts
type Posts []Post
