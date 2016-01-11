package main

import (
	"bufio"
	"github.com/russross/blackfriday"
	"gopkg.in/yaml.v2"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Page struct {
	Info PageInfo
	Body template.HTML
}

type PageInfo struct {
	Title string
	Blurb string
	Date  string
	Path  string
}

type TFile struct {
	Content  PageInfo
	LastSeen string // last time the file was changed in time.RFC1123
}

type Config struct {
	TmplDir     string // Directory containing Go html templates
	TmplExt     string // Templates should end with this extension
	BaseTmpl    string // Master template. Should be in TmplDir
	PostTmpl    string // Template used for posts
	IndexTmpl   string // Template for /
	BlogTmpl    string // Template used for list of posts
	PostExt     string // Posts should end with this extension
	ContentDir  string // Posts and static pages go here
	OutDir      string // Processed files are placed in this directory
	DumpFile    string // Information from previous runs is stored here
	ShouldWatch bool   // Watch the content directory for changes
}

type Generator struct {
	Settings     Config // see config.yml
	SeenChange   bool
	TrackedFiles map[string]TFile // filename -> file metadata
}

// NewGenerator loads settings from the specified configuration file and may
// remove all  previously generated files. Returns an error if the config file
// couldn't be opened or if the file couldn't be marshaled (most likely due to a
// YAML syntax error).
func NewGenerator(configName string, shouldClean bool) (Generator, error) {
	config, err := ioutil.ReadFile(configName)
	var g Generator
	if err != nil {
		return g, err
	}
	//var c Config
	if err = yaml.Unmarshal(config, &g.Settings); err != nil {
		return g, err
	}
	g.TrackedFiles = make(map[string]TFile)
	g.SeenChange = true
	if shouldClean {
		os.RemoveAll(g.Settings.OutDir)
		os.Remove(g.Settings.DumpFile)
	} else {
		// load dump file from previous run
		dump, ferr := ioutil.ReadFile(g.Settings.DumpFile)
		if ferr == nil {
			yaml.Unmarshal(dump, &g.TrackedFiles)
		} else {
			// if unable to load the dump file, re-render all the files
			os.RemoveAll(g.Settings.OutDir)
		}
	}
	return g, err
}

// FirstRun calls generate for the first time and sets a ticker to call generate
// every 5 seconds (unless the user chose not to watch the content directory).
func (g *Generator) FirstRun() {
	g.generate()
	if g.Settings.ShouldWatch {
		ticker := time.NewTicker(time.Second * 5)
		// BUG: if large amounts of new posts are processed each run
		// there's a chance a call to generate might not finish before the
		// ticker calls again. Fix: Put a mutex on each output file?
		go func() {
			for range ticker.C {
				g.generate()
			}
		}()
	}
}

type NoResourceError struct {
	Request string
}

func (e NoResourceError) Error() string {
	return "No mapping between " + e.Request + " and a generated file"
}

// FindResource finds the filename of the requested resource.
// After blog a gogo's first run, the root directory might look like this
// content/
//		about.tmpl
//		post/
//			first.post
//			good.post
// out/
// 		about.html
//		post/
//			first.html
//			good.html
// and a GET for http://site.com/about should return out/about.html.
// The caller should strip any leading or trailing "/" and the hostname
// (http.HandleFunc does this by default).
func (g *Generator) FindResource(res string) (string, error) {
	// if a file isn't tracked, but does exist (e.g., images), simply return it
	if doesExist(res) {
		return res, nil
	}
	if res == "" {
		res = "index"
	}
	out := g.Settings.OutDir + res + ".html"
	if !doesExist(out) {
		return out, NoResourceError{Request: res}
	}
	return out, nil
}

// renderTemplate assumes the existence of a base template and that all other
// templates implement whatever {{ define }} blocks are in the base template.
func (g *Generator) renderTemplate(w io.Writer, filename string, v interface{}) {
	base := g.Settings.TmplDir + g.Settings.BaseTmpl
	t := template.Must(template.ParseFiles(filename, base))
	if err := t.ExecuteTemplate(w, "base", v); err != nil {
		log.Fatal(err)
	} else {
		log.Print("Rendered " + filename)
	}
}

// splitFile breaks a given file up into front matter (a YAML block at the start
// of the file) and the remaining file contents.
// BUG: the actual terminator for a YAML block is "..."
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

// getOutputFilename returns the output filename for the given parameter
func (g *Generator) getOutputFilename(filename string) string {
	newPath := strings.TrimPrefix(filename, g.Settings.ContentDir)
	newPath = strings.TrimSuffix(newPath, g.Settings.TmplExt)
	newPath = strings.TrimSuffix(newPath, g.Settings.PostExt)
	return g.Settings.OutDir + newPath + ".html"
}

// stripDirectories returns everything after the last / in the parameter
func stripDirectories(filename string) string {
	path := strings.Split(filename, "/")
	return path[len(path)-1]
}

// processFile renders the Markdown for a post and calls renderTemplate() with
// the post template OR simply calls renderTemplate() if the filename ends with
// the template extension.
func (g *Generator) processFile(filename string) Page {
	frontMatter, content := splitFile(filename)
	var info PageInfo
	yaml.Unmarshal([]byte(frontMatter), &info)
	p := Page{info, ""}
	finalFilename := g.getOutputFilename(filename)
	output, err := os.Create(finalFilename)
	defer output.Close()
	if err != nil {
		log.Fatal(err)
	}
	if strings.Contains(filename, g.Settings.PostExt) {
		// parse markdown, render post template
		p.Body = template.HTML(blackfriday.MarkdownCommon([]byte(content)))
		p.Info.Path = g.getOutputFilename(filename)
		p.Info.Path = strings.Replace(p.Info.Path, ".html", "", 1)
		p.Info.Path = strings.Replace(p.Info.Path, g.Settings.OutDir, "", 1)
		g.renderTemplate(output, g.Settings.TmplDir+g.Settings.PostTmpl, p)

	} else {
		// render template
		tmpPrefix := stripDirectories(filename)
		tmpFile, err := ioutil.TempFile("", tmpPrefix)
		if err != nil {
			log.Fatal(err)
		}
		defer tmpFile.Close()
		if err := ioutil.WriteFile(tmpFile.Name(), []byte(content), 0755); err != nil {
			log.Fatal(err)
		}
		g.renderTemplate(output, tmpFile.Name(), p)
		os.Remove(tmpFile.Name())
	}
	return p
}

// isUnchanged returns true if the modification time of the file matches the
// time it was first recorded
func (g *Generator) isUnchanged(filename string, info os.FileInfo) bool {
	trackedFile, ok := g.TrackedFiles[filename]
	modTime := info.ModTime().Format(time.RFC1123)
	return ok && (trackedFile.LastSeen == modTime)
}

// isUnsupported checks if the filename contains a post or template extension
func (g *Generator) isUnsupported(path string, info os.FileInfo) bool {
	isTemplate := strings.Contains(path, g.Settings.TmplExt)
	isPost := strings.Contains(path, g.Settings.PostExt)
	return !(isTemplate || isPost)
}

// makeVisit returns a function that filepath.Walk can call.
func (g *Generator) makeVisit() func(string, os.FileInfo, error) error {
	// this anonymous function is called by filepath.Walk and does 1 of 3 things
	// - if there is a directory in contentDir, recreate it in outDir
	// - if a file is unchanged or unsupported, ignore it
	// - otherwise, render it
	// TODO: fire off each call to processFile in its own goroutine
	return func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if _, err := os.Stat(g.Settings.OutDir + info.Name()); os.IsNotExist(err) {
				err = os.Mkdir(g.Settings.OutDir+info.Name(), 0755)
				if err != nil {
					log.Print(err)
				}
			}
			return nil
		}
		if g.isUnsupported(path, info) || g.isUnchanged(path, info) {
			return nil
		}
		p := g.processFile(path)
		g.TrackedFiles[path] = TFile{p.Info, info.ModTime().Format(time.RFC1123)}
		g.SeenChange = true
		return nil
	}
}

// generate simply begins the walk through the filesystem and dumps the results
// into a file if there have been any changes
func (g *Generator) generate() {
	if _, err := os.Stat(g.Settings.OutDir); os.IsNotExist(err) {
		err := os.Mkdir(g.Settings.OutDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
	filepath.Walk(g.Settings.ContentDir, g.makeVisit())
	if g.SeenChange {
		g.makeBlog()
		g.SeenChange = false
		info, _ := yaml.Marshal(&g.TrackedFiles)
		ioutil.WriteFile(g.Settings.DumpFile, info, 0664)
	}
}

// doesExist is a wrapper around os.Stat(filename). Technically, the return
// value should be os.IsExist(err), but blog a gogo can't handle any filesystem
// errors that os.Open() or os.Create() might return, so checking for nil is
// the best that can be done
func doesExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// makeBlog calls renderTemplate with a list of posts and the blog template
// TODO: sort posts by date
func (g *Generator) makeBlog() {
	pages := make([]PageInfo, 0)
	for filename, p := range g.TrackedFiles {
		if strings.Contains(filename, g.Settings.PostExt) {
			pages = append(pages, p.Content)
		}
	}
	finalFilename := g.getOutputFilename(g.Settings.BlogTmpl)
	if doesExist(finalFilename) {
		os.Remove(finalFilename)
	}
	f, e := os.Create(finalFilename)
	if e != nil {
		log.Fatal(e)
	}
	defer f.Close()
	g.renderTemplate(f, g.Settings.TmplDir+g.Settings.BlogTmpl, pages)
}
