package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"
)

var (
	b *Blogger
	// initialization constants
	STATIC_DIR string = "/static/"
	TMPL_DIR   string = "/tmpl/"
	POST_DIR   string = "/post/"
	POST_FN    string = "posts.json"
	ROUTES_FN  string = "routes.json"
	SITE_URL   string = "localhost/"
)

type Session struct {
	ExpirDate time.Time
	// attempts?
}

type Resources struct {
	StaticDir string
	TmplDir   string
	PostDir   string
}

type Blogger struct {
	Site           *Blog
	ActiveSessions map[string]Session // r.RemoteAddr -> Session
	Paths          Resources
	Answer         string
}

func NewBlogger(answer string) *Blogger {
	paths := Resources{StaticDir: STATIC_DIR, TmplDir: TMPL_DIR, PostDir: POST_DIR}
	sessions := make(map[string]Session)
	blag := NewBlog(POST_FN, SITE_URL)
	return &Blogger{Site: blag, ActiveSessions: sessions, Paths: paths, Answer: answer}
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[len(b.Paths.StaticDir):]
	if filename == "" {
		http.NotFound(w, r)
		return
	}
	f, err := http.Dir(b.Paths.StaticDir[1:]).Open(filename)
	if err != nil {
		fmt.Println(filename + " does not exist")
	} else {
		file := io.ReadSeeker(f)
		http.ServeContent(w, r, filename, time.Now(), file)
	}
}

// /about -> tmpl/about.html

func pageHandler(w http.ResponseWriter, r *http.Request) {
	filename := b.Site.Routes[r.URL.Path[1:]]
	fmt.Println(filename + " requested")
	if filename == "" {
		fmt.Println("Unmapped reource " + filename)
		http.NotFound(w, r)
		return
	}
	t, err := template.ParseFiles(filename)
	if err != nil {
		panic(err)
	}
	if err = t.Execute(w, b); err != nil {
		fmt.Println(err)
	}
}

// HOWTO use Blag w/o configuration
// import "blag"
// func main() {
//  blag.Run(blag.NewBlogger("type me if you want to edit"), ":80")
//}

func Run(blag *Blogger, port string) {
	b = blag
	http.HandleFunc(b.Paths.StaticDir, staticHandler)
	http.HandleFunc("/", pageHandler)
	http.ListenAndServe(port, nil)
}

func main() {
	blag := NewBlogger("top kek")
	Run(blag, ":80")
}
