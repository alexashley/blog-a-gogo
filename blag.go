package main

import (
	"flag"
	"html/template"
	"io"
	"log"
	//"net/http"
)

var (
	cssDir     = flag.String("cssDir", "css/", "Directory containing css files")
	tmplDir    = flag.String("tmplDir", "tmpl/", "Directory containing Go html templates")
	baseTmpl   = flag.String("base", "base.tmpl", "Master template. Should be in tmplDir")
	postTmpl   = flag.String("post", "post.tmpl", "Template to be used for blog posts")
	contentDir = flag.String("contentDir", "content/", "Posts & static pages go here")
	outDir     = flag.String("outDir", "out/", "Processed pages are output here")
	watch      = flag.Bool("watch", true, "Watch contentDir for changes")
	host       = flag.String("host", "localhost", "Hostname")
	port       = flag.String("port", ":8080", "Port for http.ListenAndServe")
)

type Page struct {
	Title  string
	Body   []byte
	URL    string
	cssDir string
	Date   string
}

func renderTemplate(w io.Writer, filename string, p Page) {
	t := template.Must(template.ParseFiles(filename, *tmplDir+*baseTmpl))
	if err := t.ExecuteTemplate(w, "base", p); err != nil {
		log.Fatal("Error parsing template " + filename)
	}
}

// generatePosts processes the files in contentDir and places the HTML in outDir
// *.md -> html -> rendered w/ postTmpl (inherits from baseTmpl)
// .tmpl -> rendered w/ baseTmpl
func generateSite() {

}

func main() {
	flag.Parse()

}
