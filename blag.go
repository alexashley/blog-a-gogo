package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"time"
)

type Post struct {
	Title string
	Body  template.HTML
	Blurb string
	Date  time.Time
}

type Blog struct {
	Posts  map[string]Post   // postID -> post struct
	Routes map[string]string // about -> tmpl/about.html
	URL    string
}

func NewBlog(resFile string, postFile string, url string) *Blog {
	posts := make(map[string]Post)
	routes := make(map[string]string)
	LoadFromJSON(postFile, &posts)
	LoadFromJSON(resFile, &routes)
	return &Blog{Posts: posts, Routes: routes, URL: url}
}

func (b *Blog) SavePosts(filename string) {
	data, _ := json.MarshalIndent(b.Posts, "", "\t")
	ioutil.WriteFile(filename, data, 0644)
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
