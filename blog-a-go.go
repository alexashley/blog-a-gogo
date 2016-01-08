package main

import (
	"bufio"
	"flag"
	"github.com/russross/blackfriday"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	//"net/http"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
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
	Info   PageInfo
	Body   []byte
	URL    string
	CSSDir string
}

type PageInfo struct {
	Title string
	Blurb string
	Date  string
}

type WatchedFile struct {
	lastSeen time.Time
}

func renderTemplate(w io.Writer, filename string, p Page) {
	t := template.Must(template.ParseFiles(filename, *tmplDir+*baseTmpl))
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
	newPath := strings.TrimPrefix(filename, *contentDir)
	newPath = strings.TrimSuffix(newPath, ".tmpl")
	newPath = strings.TrimSuffix(newPath, ".post")
	return *outDir + newPath + ".html"
}

func processFile(filename string) {
	frontMatter, content := splitFile(filename)
	var info PageInfo
	yaml.Unmarshal([]byte(frontMatter), &info)
	var b []byte
	p := Page{info, b, *host, *cssDir}
	finalFilename := getOutputFilename(filename)
	output, err := os.Create(finalFilename)
	defer output.Close()
	if err != nil {
		log.Fatal(err)
	}
	if strings.Contains(filename, ".post") {
		// parse markdown
		p.Body = blackfriday.MarkdownCommon([]byte(content))
		renderTemplate(output, *postTmpl, p)
		//ioutil.WriteFile(finalFilename, html, 0755)

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
		/*finalFilename := getOutputFilename(filename)
		output, err := os.Create(finalFilename)
		if err != nil {
			log.Fatal(err)
		}*/
		renderTemplate(output, tmpFile.Name(), p)
		//output.Close()
		os.Remove(tmpFile.Name())
	}
}

func isUnchanged(filename string, info os.FileInfo) bool {
	f, ok := allFiles[filename]
	return ok && f.lastSeen.Equal(info.ModTime())
}

func isUnsupported(path string, info os.FileInfo) bool {
	isTemplate := strings.Contains(path, ".tmpl")
	isPost := strings.Contains(path, ".post")
	return !(isTemplate || isPost)
}

func walkFn(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		if _, err := os.Stat(*outDir + path); os.IsNotExist(err) {
			err = os.Mkdir(*outDir+info.Name(), 0755)
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

// generatePosts processes the files in contentDir and places the HTML in outDir
// *.md -> html -> rendered w/ postTmpl (inherits from baseTmpl)
// .tmpl -> rendered w/ baseTmpl
func generateSite() {
	err := os.Mkdir(*outDir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	filepath.Walk(*contentDir, walkFn)
}

func main() {
	flag.Parse()
	allFiles = make(map[string]WatchedFile)
	//	os.RemoveAll(*outDir)
	generateSite()
}
