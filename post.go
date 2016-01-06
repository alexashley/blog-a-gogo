package main

import (
	"html/template"
	"time"
)

type Post struct {
	Title    string
	Filename string
	Blurb    string // brief post summary
	Date     time.Time
	LastEdit time.Time
	Body     template.HTML
}
