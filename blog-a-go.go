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
	config    = flag.String("config", "config.yaml", "Yaml configuration file")
	watch     = flag.Bool("watch", true, "Watch contentDir for changes")
	runServer = flag.Bool("runServer", true, "Run server on -host, listening on -port")
	host      = flag.String("host", "http://localhost:8080/", "Hostname.")
	port      = flag.String("port", ":8080", "Port for http.ListenAndServe")

	allFiles map[string]WatchedFile
	settings Config
)

type Page struct {
	Info PageInfo
	Body []byte
	URL  string
}

type PageInfo struct {
	Title string
	Blurb string
	Date  string
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

type WatchedFile struct {
	lastSeen time.Time
}

func renderTemplate(w io.Writer, filename string, p Page) {
	t := template.Must(template.ParseFiles(filename, settings.TmplDir+settings.BaseTmpl))
	if err := t.ExecuteTemplate(w, "base", p); err != nil {
		log.Fatal(err)
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

func processFile(filename string) {
	frontMatter, content := splitFile(filename)
	var info PageInfo
	yaml.Unmarshal([]byte(frontMatter), &info)
	var b []byte
	p := Page{info, b, *host}
	finalFilename := getOutputFilename(filename)
	output, err := os.Create(finalFilename)
	defer output.Close()
	if err != nil {
		log.Fatal(err)
	}
	if strings.Contains(filename, settings.PostExt) {
		// parse markdown
		p.Body = blackfriday.MarkdownCommon([]byte(content))
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
}

func isUnchanged(filename string, info os.FileInfo) bool {
	f, ok := allFiles[filename]
	return ok && f.lastSeen.Equal(info.ModTime())
}

func isUnsupported(path string, info os.FileInfo) bool {
	isTemplate := strings.Contains(path, settings.TmplExt)
	isPost := strings.Contains(path, settings.PostExt)
	return !(isTemplate || isPost)
}

func walkFn(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		if _, err := os.Stat(settings.OutDir + path); os.IsNotExist(err) {
			err = os.Mkdir(settings.OutDir+info.Name(), 0755)
			if err != nil {
				log.Print(err)
			}
		}
		return nil
	}
	// unchanged and unsupported files
	if isUnsupported(path, info) || isUnchanged(path, info) {
		return nil
	}
	processFile(path)
	allFiles[path] = WatchedFile{lastSeen: info.ModTime()}

	return nil
}

func generateSite() {
	err := os.Mkdir(settings.OutDir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	filepath.Walk(settings.ContentDir, walkFn)
}

func handler(w http.ResponseWriter, r *http.Request) {
	filename := strings.Trim(r.URL.Path, "/") + ".html"
	f, err := http.Dir(settings.OutDir).Open(filename)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	file := io.ReadSeeker(f)
	http.ServeContent(w, r, filename, time.Now(), file)
}

// get a list o' posts and display 'em
func generateBlog() {

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
	allFiles = make(map[string]WatchedFile)
	// blog-a-gogo has been run before. Load previous state of allFiles
	if _, err := os.Stat(settings.OutDir); os.IsExist(err) {
		info, err := ioutil.ReadFile(settings.WatchFile)
		// ignore any errors, just rebuild allFiles and templates
		if err == nil {
			yaml.Unmarshal(info, &allFile)
		} else {
			os.RemoveAll(settings.OutDir)
		}
	}

	generateSite()
	generateBlog()
	if *runServer {
		http.HandleFunc("/", handler)
		http.ListenAndServe(*port, nil)
	}
}
