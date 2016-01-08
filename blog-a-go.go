package main

import (
	"flag"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	//"net/http"
	"strings"
	"time"
)

var (
	cssDir     = flag.String("cssDir", "css/", "Directory containing css files")
	tmplDir    = flag.String("tmplDir", "tmpl/", "Directory containing Go html templates")
	baseTmpl   = flag.String("base", "base.tmpl", "Master template.")
	postTmpl   = flag.String("post", "post.tmpl", "Template to be used for blog posts")
	contentDir = flag.String("contentDir", "content/", "Posts & static pages go here")
	outDir     = flag.String("outDir", "out/", "Processed pages are output here")
	watch      = flag.Bool("watch", true, "Watch contentDir for changes")
	host       = flag.String("host", "localhost", "Hostname")
	port       = flag.String("port", ":8080", "Port for http.ListenAndServe")

	allFiles map[string]WatchedFile
)

type Page struct {
	Title  string
	Body   []byte
	URL    string
	cssDir string
	Date   string
}

type WatchedFile struct {
	lastSeen time.Time
}

func renderTemplate(w io.Writer, filename string, p Page) {
	t := template.Must(template.ParseFiles(filename, *tmplDir+*baseTmpl))
	if err := t.ExecuteTemplate(w, "base", p); err != nil {
		log.Fatal("Error parsing template " + filename)
	}
}

func renderHTML(b []byte) {

}

//
func processFile(filename string) {
	var p Page
	p.URL = *host
	p.cssDir = *cssDir
}

func isUnchanged(filename string, info os.FileInfo) bool {
	f, ok := allFiles[filename]
	return ok && f.lastSeen.Equal(info.ModTime())
}

func isUnsupported(path string, info os.FileInfo) bool {
	isTemplate := strings.Contains(path, ".tmpl")
	isMarkdown := strings.Contains(path, ".md")
	return !(isTemplate || isMarkdown) || info.IsDir()
}

func walkFn(path string, info os.FileInfo, err error) error {
	// ignore directories and unchanged files
	if isUnsupported(path, info) || isUnchanged(path, info) {
		return nil
	}
	processFile(path)
	allFiles[path] = WatchedFile{lastSeen: info.ModTime()}

	return nil
}

// generatePosts processes the files in contentDir and places the HTML in outDir
// *.md -> html -> rendered w/ postTmpl (inherits from baseTmpl)
// .tmpl -> rendered w/ baseTmpl
func generateSite() {
	filepath.Walk(*contentDir, walkFn)
}

func main() {
	flag.Parse()
	allFiles = make(map[string]WatchedFile)
}
