package blag

import (
	"html/template"
	"time"
)

type Post struct {
	Title    string
	Filename string
	Blurb    string // brief summary of post
	Date     time.Time
	Body     template.HTML
}

// LoadPosts dumps post metadata from the specified file
// TODO: sort by date. posts[0] should be the most recent
func LoadPosts(fn string) []Post {
	return make([]Post, 10)
}
