package main

import (
	"net/http"
)

var b *Blog

func staticHandler(w http.ResponseWriter, r *http.Request) {

}

func pageHandler(w http.ResponseWriter, r *http.Request) {

}

// TODO: read options from stdin
// switches:
// 	-p filename w/ all post metadata in JSON format
// 	-c filename containing non-vanilla configuration
func main() {
	b = NewBlog("")
	http.HandleFunc(b.Paths.StaticDir, staticHandler)
	http.HandleFunc("/", pageHandler)
	http.ListenAndServe(b.Port, nil)
}
