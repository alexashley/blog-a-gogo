package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	config    = flag.String("config", "config.yml", "YAML configuration file")
	clean     = flag.Bool("clean", false, "Remove all files from previous runs.")
	port      = flag.String("port", "8080", "port number")
	runServer = flag.Bool("runServer", false, "Run server on -port")
	site      Generator
)

func handler(w http.ResponseWriter, r *http.Request) {
	req := strings.Trim(r.URL.Path, "/")
	fname, resErr := site.FindResource(req)
	if resErr != nil {
		log.Print("Unknown resource " + req + " requested")
		http.NotFound(w, r)
		return
	}
	log.Print("file " + req + " requested and " + fname + " found")
	f, err := http.Dir("").Open(fname)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	file := io.ReadSeeker(f)
	http.ServeContent(w, r, fname, time.Now(), file)
}

func init() {
	flag.Parse()
	var err interface{}
	site, err = NewGenerator(*config, *clean)
	if err != nil {
		log.Fatal(err)
	}
	site.FirstRun()
}

func main() {
	if *runServer {
		http.HandleFunc("/", handler)
		http.ListenAndServe(":"+*port, nil)
	}
}
