package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"
)

var (
	blog      *Blog
	staticDir string
)

func staticHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[len(staticDir):]
	if filename == "" {
		http.NotFound(w, r)
		return
	}
	f, err := http.Dir(staticDir[1:]).Open(filename)
	if err != nil {
		fmt.Println(filename + " does not exist")
	} else {
		file := io.ReadSeeker(f)
		http.ServeContent(w, r, filename, time.Now(), file)
	}
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	// check for empty string
	filename := blog.Routes[r.URL.Path[1:]]
	if filename == "" {
		fmt.Println("Unmapped resource " + r.URL.Path)
		http.NotFound(w, r)
		return
	}
	fmt.Println(filename)
	t, err := template.ParseFiles(filename)
	if err != nil {
		panic(err)
	}
	if err = t.Execute(w, blog); err != nil {
		fmt.Println(err)
	}
}

// TODO: read options from stdin
// switches:
// 	-p filename w/ all post metadata in JSON format
//  -s path to static directory
//  -r route file
func main() {
	blog = NewBlog("routes.json", "posts.json", "localhost/")
	staticDir = "/static/"
	http.HandleFunc(staticDir, staticHandler)
	http.HandleFunc("/", pageHandler)
	http.ListenAndServe(":80", nil)
}
