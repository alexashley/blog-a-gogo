package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"time"
)

type Post struct {
	Title    string
	Body     template.HTML
	Filename string
	Blurb    string
	Date     time.Time
}

type Blog struct {
	Posts  []Post            // post objects in memory
	Routes map[string]string // about/ -> tmpl/about.html
	URL    string
}

func NewBlog(resFile string, postFile string, url string) *Blog {
	posts := make([]Post, 0)
	routes := make(map[string]string)
	LoadFromJSON(postFile, &posts)
	LoadFromJSON(resFile, &routes)
	return &Blog{Posts: posts, Routes: routes, URL: url}
}

func LoadFromJSON(filename string, obj interface{}) {
	j, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Unable to read from " + filename)
	}
	if err = json.Unmarshal(j, obj); err != nil {
		fmt.Println("Error parsing JSON from " + filename)
	}
}
