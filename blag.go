package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	PORT_NUM   string = ":80" // default http port
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
	Posts  []Post            // post objects in memory TODO: sort by date
	Port   string
}

func NewBlog(filename string) *Blog {
	var paths Resources
	if filename != "" {
		LoadFromJSON(filename, &paths)
	} else {
		paths = Resources{StaticDir: STATIC_DIR,
			TmplDir:  TMPL_DIR,
			BaseTmpl: BASE_TMPL,
			ErrorDir: ERROR_DIR,
			PostDir:  POST_DIR}
	}
	routes := make(map[string]string)
	posts := make([]Post, 0)
	LoadFromJSON(ROUTES_FN, &routes)
	LoadFromJSON(POST_FN, &posts)

	return &Blog{Paths: paths, Routes: routes, Posts: posts, Port: PORT_NUM}
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
