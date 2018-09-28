package main

import (
	"time"
)

// Post contains all data for one blog post
type Post struct {
	ID      string    `json:"id"`
	IsShort bool      `json:"isshort"`
	Title   string    `json:"title"`
	Visible bool      `json:"visible"`
	Date    time.Time `json:"date"`
	Body    string    `json:"body"`
}

// Posts is just an array of posts
type Posts []Post
