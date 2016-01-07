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
	Posts     []Post            // post objects in memory
	Resources map[string]string // about/ -> tmpl/about.html
	URL       string
}

func NewBlog(resFile string, postFile string, url string) *Blog {
	posts := make([]Post, 0)
	resources := make(map[string]string)
	LoadFromJSON(postFile, &posts)
	LoadFomJSON(resFile, &resources)
	return &Blog{Posts: posts, Resources: resources, URL: url}
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
