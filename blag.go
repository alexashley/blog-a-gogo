package blag

import (
	"net/http"
)

var (
	// default values for Resources struct
	STATIC_DIR string = "/static/"
	TMPL_DIR   string = "/tmpl/"
	BASE_TMPL  string = "/tmpl/base.html"
	ERROR_DIR  string = "/tmpl/error"
	POST_DIR   string = "/post/"
	POST_FN    string = "posts.json" // post metadata in json format
	ROUTES_FN  string = "routes.json"
	PORT_NUM   int    = 80 // default http port
)

type Resources struct {
	StaticDir string // static files (e.g., css, js)
	TmplDir   string // templates
	BaseTmpl  string // site-wide template
	ErrorDir  string // custom error responses. default: /tmpl/error/
	PostDir   string // actual post files (extensions of the base template)
}

type Blog struct {
	Paths  Resources
	Routes map[string]string // maps URLs to files, e.g., site.com/about -> tmpl/site.html
	Posts  []Post            // post objects in memory
	Port   int
}

func New() Blog {
	paths := Resources{StaticDir: STATIC_DIR,
		TmplDir:  TMPL_DIR,
		BaseTmpl: BASE_TMPL,
		ErrorDir: ERROR_DIR,
		PostDir:  POST_DIR}

	routes := LoadRoutes(ROUTES_FN)
	posts := LoadPosts(POST_FN)

	return Blog{Paths: paths, Routes: routes, Posts: posts, Port: PORT_NUM}
}

func LoadRoutes(fn string) map[string]string {
	return make(map[string]string)
}

func (b *Blog) Run() {

}

// serves files from the static directory
func staticHandler(w http.ResponseWriter, r *http.Request) {

}

// serves custom error pages (e.g., 404, 503)
func errorHandler(w http.ResponseWriter, r *http.Request) {

}
