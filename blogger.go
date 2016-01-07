package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var (
	b *Blogger
	// initialization constants
	STATIC_DIR string = "/static/"
	/*TMPL_DIR   string = "tmpl/"*/
	BASE_TMPL string = "tmpl/base.html"
	POST_DIR  string = "posts/"
	POST_FN   string = "posts.json"
	ROUTES_FN string = "routes.json"
	SITE_URL  string = "http://localhost/"
)

type Session struct {
	ExpirDate time.Time
	// attempts?
}

type Resources struct {
	StaticDir string
	BaseTmpl  string
	PostDir   string
}

type Blogger struct {
	Site           *Blog
	CurrID         string             // post id
	ActiveSessions map[string]Session // r.RemoteAddr -> Session
	Paths          Resources
	Answer         string
}

func NewBlogger(answer string) *Blogger {
	paths := Resources{StaticDir: STATIC_DIR, BaseTmpl: BASE_TMPL, PostDir: POST_DIR}
	sessions := make(map[string]Session)
	blag := NewBlog(ROUTES_FN, POST_FN, SITE_URL)
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

func renderTemplate(w http.ResponseWriter, filename string) {
	t := template.Must(template.ParseFiles(filename, b.Paths.BaseTmpl))

	if err := t.ExecuteTemplate(w, "base", b); err != nil {
		fmt.Println(err)
	}
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	filename := b.Site.Routes[r.URL.Path[1:]]
	fmt.Println(filename + " requested")
	if filename == "" {
		fmt.Println("Unmapped resource " + filename)
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, filename)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	title := r.PostFormValue("title")
	blurb := r.PostFormValue("blurb")
	content := []byte(r.PostFormValue("content"))
	fmt.Println("POST TITLE: " + title)
	t := time.Now()
	p := Post{Title: title, Body: content, Blurb: blurb, Date: t}
	postID := strconv.FormatInt(t.Unix(), 10)
	b.Site.Posts[postID] = p
	fn := b.Paths.PostDir + postID + ".txt"
	fmt.Println("Saving post " + fn)
	ioutil.WriteFile(fn, content, 0666 /* ooh spoky! */)
	b.Site.SavePosts(POST_FN)
	http.Redirect(w, r, "/post/"+postID, http.StatusFound)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	b.CurrID = r.URL.Path[len("/post/"):]
	// := b.Site.Posts[postID]
	renderTemplate(w, b.Site.Routes["post"])
}

// url/new
func newHandler(w http.ResponseWriter, r *http.Request) {
	/*if !hasCredentials(r.RemoteAddr) {
		http.Redirect(w, r, "/authenticate", http.StatusFound)
	}*/
	renderTemplate(w, b.Site.Routes["new"])
}

func hasCredentials(address string) bool {
	session := b.ActiveSessions[address]
	if time.Now().Before(session.ExpirDate) {
		return true
	} else {
		delete(b.ActiveSessions, address)
	}
	return false
}

// HOWTO use Blag w/o configuration
// import "blag"
// func main() {
//  blag.Run(blag.NewBlogger("type me if you want to edit"), ":80")
//}

func Run(blag *Blogger, port string) {
	b = blag
	http.HandleFunc(b.Paths.StaticDir, staticHandler)
	http.HandleFunc("/new/", newHandler)
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/post/", postHandler)
	http.HandleFunc("/", pageHandler)
	http.ListenAndServe(port, nil)
}

func main() {
	blag := NewBlogger("top kek")
	Run(blag, ":80")
}
