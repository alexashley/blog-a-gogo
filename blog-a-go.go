package main

import (
	"bufio"
	"flag"
	"github.com/russross/blackfriday"
	"gopkg.in/yaml.v2"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	config    = flag.String("config", "config.yml", "Yaml configuration file")
	watch     = flag.Bool("watch", true, "Watch contentDir for changes")
	runServer = flag.Bool("runServer", true, "Run server on -host, listening on -port")
	host      = flag.String("host", "http://localhost:8080/", "Hostname.")
	port      = flag.String("port", ":8080", "Port for http.ListenAndServe")

	allFiles map[string]TrackedFile
	settings Config
)

type Page struct {
	Info PageInfo
	Body template.HTML
	URL  string
}

type PageInfo struct {
	Title string
	Blurb string
	Date  string
}

type TrackedFile struct {
	Content  Page
	LastSeen string
}

type Config struct {
	TmplDir    string // Directory containing Go html templates
	TmplExt    string // Templates should end with this extension
	BaseTmpl   string // Master template. Should be in TmplDir
	PostTmpl   string // Template used for posts
	BlogTmpl   string // Template used for list of posts
	PostExt    string // Posts should end with this extension
	ContentDir string // Posts and static pages go here
	OutDir     string // Processed files are placed in this directory
	WatchFile  string // Information from previous runs is stored here
}

func renderTemplate(w io.Writer, filename string, v interface{}) {
	base := settings.TmplDir + settings.BaseTmpl
	t := template.Must(template.ParseFiles(filename, base))
	if err := t.ExecuteTemplate(w, "base", v); err != nil {
		log.Fatal(err)
	} else {
		log.Print("Rendered " + filename)
	}
}

func splitFile(filename string) (frontMatter string, content string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	seenFM := true
	line := ""
	for scanner.Scan() {
		line = scanner.Text() + "\n"
		if strings.TrimSpace(line) == "" {
			continue
		}
		if scanner.Text() == "---" {
			seenFM = !seenFM
			continue
		}
		if seenFM {
			content += line
		} else {
			frontMatter += line
		}
	}
	return frontMatter, content
}

func getOutputFilename(filename string) string {
	newPath := strings.TrimPrefix(filename, settings.ContentDir)
	newPath = strings.TrimSuffix(newPath, settings.TmplExt)
	newPath = strings.TrimSuffix(newPath, settings.PostExt)
	return settings.OutDir + newPath + ".html"
}

func processFile(filename string) Page {
	frontMatter, content := splitFile(filename)
	var info PageInfo
	yaml.Unmarshal([]byte(frontMatter), &info)
	p := Page{info, "", *host}
	finalFilename := getOutputFilename(filename)
	output, err := os.Create(finalFilename)
	defer output.Close()
	if err != nil {
		log.Fatal(err)
	}
	if strings.Contains(filename, settings.PostExt) {
		// parse markdown
		p.Body = template.HTML(blackfriday.MarkdownCommon([]byte(content)))
		renderTemplate(output, settings.TmplDir+settings.PostTmpl, p)

	} else {
		// render template
		path := strings.Split(filename, "/")
		tmpPrefix := path[len(path)-1]
		tmpFile, err := ioutil.TempFile("", tmpPrefix)
		if err != nil {
			log.Fatal(err)
		}
		defer tmpFile.Close()
		if err := ioutil.WriteFile(tmpFile.Name(), []byte(content), 0755); err != nil {
			log.Fatal(err)
		}
		renderTemplate(output, tmpFile.Name(), p)
		os.Remove(tmpFile.Name())
	}
	return p
}

func isUnchanged(filename string, info os.FileInfo) bool {
	trackedFile, ok := allFiles[filename]
	modTime := info.ModTime().Format(time.RFC1123)
	return ok && (trackedFile.LastSeen == modTime)
}

func isUnsupported(path string, info os.FileInfo) bool {
	isTemplate := strings.Contains(path, settings.TmplExt)
	isPost := strings.Contains(path, settings.PostExt)
	return !(isTemplate || isPost)
}

func walkFn(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		if _, err := os.Stat(settings.OutDir + info.Name()); os.IsNotExist(err) {
			err = os.Mkdir(settings.OutDir+info.Name(), 0755)
			if err != nil {
				log.Print(err)
			}
		}
		return nil
	}
	if isUnsupported(path, info) || isUnchanged(path, info) {
		return nil
	}
	p := processFile(path)
	allFiles[path] = TrackedFile{p, info.ModTime().Format(time.RFC1123)}

	return nil
}

func generateSite() {
	if _, err := os.Stat(settings.OutDir); os.IsNotExist(err) {
		err := os.Mkdir(settings.OutDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
	filepath.Walk(settings.ContentDir, walkFn)
	info, _ := yaml.Marshal(&allFiles)
	ioutil.WriteFile(settings.WatchFile, info, 0664)
	generateBlog()
}

func doesExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	filename := strings.Trim(r.URL.Path, "/")
	outf := settings.OutDir + filename + ".html"
	// if the file is tracked, load it from the out directory
	if doesExist(outf) {
		filename = outFilename
	}
	log.Print(filename + " requested")
	f, err := http.Dir("").Open(filename)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	file := io.ReadSeeker(f)
	http.ServeContent(w, r, filename, time.Now(), file)
}

type Blog struct {
	Pages []Page
	URL   string
}

// get a list o' posts and organize them
func generateBlog() {
	pages := make([]Page, 0)
	for filename, p := range allFiles {
		if strings.Contains(filename, settings.PostExt) {
			pages = append(pages, p.Content)
		}
	}
	finalFilename := getOutputFilename(settings.BlogTmpl)
	if _, err := os.Stat(finalFilename); err == nil {
		os.Remove(finalFilename)
	}
	f, e := os.Create(finalFilename)
	if e != nil {
		log.Fatal(e)
	}
	defer f.Close()
	b := Blog{pages, *host}
	renderTemplate(f, settings.TmplDir+settings.BlogTmpl, b)
}

func loadConfig(filename string) {
	config, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Unable to load configuration file")
	}
	yaml.Unmarshal(config, &settings)
}

func main() {
	flag.Parse()
	loadConfig(*config)
	allFiles = make(map[string]TrackedFile)
	// blog-a-gogo has been run before. Load previous state of allFiles
	if _, err := os.Stat(settings.WatchFile); err == nil {
		info, err := ioutil.ReadFile(settings.WatchFile)
		// ignore any errors, just rebuild allFiles and templates
		if err == nil {
			yaml.Unmarshal(info, &allFiles)
		} else {
			os.RemoveAll(settings.OutDir)
		}
	}
	generateSite()
	if *runServer {
		http.HandleFunc("/", handler)
		http.ListenAndServe(*port, nil)
	}
}
