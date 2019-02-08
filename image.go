package main

import (
	"time"
)

// Image contains all data for one image
type Image struct {
	Filename string    `json:"filename"`
	Title    string    `json:"title"`
	AltText  string    `json:"alttext"`
	URL      string    `json:"url"`
	Date     time.Time `json:"date"`
}

// Images is just an array of posts
type Images []Image
