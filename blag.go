package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Blog struct {
	Routes map[string]string // maps URLs to files, e.g., site.com/about -> tmpl/site.html
	Posts  []Post            // post objects in memory TODO: sort by date
	URL    string
}

func NewBlog(routesFile string, postFile string, url string) *Blog {
	routes := make(map[string]string)
	posts := make([]Post, 0)
	LoadFromJSON(routesFile, &routes)
	LoadFromJSON(postFile, &posts)
	return &Blog{Routes: routes, Posts: posts, URL: url}
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
