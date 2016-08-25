package view

import (
	"html/template"
	"sync"
	"io"
	"path/filepath"
	"bytes"
	"regexp"
	"strconv"
	"strings"
	"bufio"
	"fmt"
)

type ParseError struct {
	// Because it is hard to find the line number from a merged file,
	// so just return a text line.
	Line  string

	// The original error message.
	Err   string

	// The error may come from multiple pages.
	Pages []Page
}

func (pe *ParseError) Error() string {
	var pages string
	for _, p := range pe.Pages {
		pages += "\n" + filepath.Join(p.Dir, p.File)
	}
	return fmt.Sprintf("Parse view file got error: %s;\nText line: %s\nError from: %s", pe.Err, pe.Line, pages)
}

type Container struct {
	tpls    map[string]*template.Template
	ext     string
	debug   bool
	handler func(tpl *template.Template)
	rwmu    *sync.RWMutex
}

// If debug is true, combiner will always read view file from disk,
// otherwise it will cache the Template objects.
func NewContainer(debug bool, ext string) *Container {

	return &Container{
		tpls: make(map[string]*template.Template, 15),
		debug: debug,
		ext: ext,
		rwmu: &sync.RWMutex{},
	}
}

func (this *Container) SetTplHandle(handler func(tpl *template.Template)) {

	this.handler = handler
}

// Clear all cache
func (this *Container) Clear() {

	this.tpls = make(map[string]*template.Template, 15)
}

type Page struct {
	Dir  string
	File string
	Debug bool
}

func NewPage(dir, file string) Page {
	return Page{
		Dir: dir,
		File: file,
	}
}

// Debug page will be ignored form errors
func NewDebugPage(dir, file string) Page {
	return Page {
		Dir: dir,
		File: file,
		Debug: true,
	}
}

func (this *Container) Combine(ps ...Page) (html []byte, err error) {

	pNum := len(ps)
	pages := make([][]byte, pNum)
	for idx, s := range ps {
		html, err = Combine(string(s.Dir), string(s.File), this.ext)
		if err != nil {
			return
		}
		pages[idx] = html
	}
	if pNum > 1 {
		html = MergeHtml(pages)
	} else if pNum == 1 {
		html = pages[0]
	}

	return
}

func (this *Container) Display(w io.Writer, data interface{}, ps ...Page) error {

	buf := bytes.NewBuffer(nil)
	for _, s := range ps {
		buf.WriteString(s.Dir)
		buf.WriteRune(filepath.Separator)
		buf.WriteString(s.File)
	}
	name := buf.String()
	this.rwmu.RLock()
	tpl, ok := this.tpls[name]
	this.rwmu.RUnlock()
	if ok {
		return tpl.Execute(w, data)
	} else {
		html, err := this.Combine(ps...)
		if err != nil {
			return err
		}
		this.rwmu.Lock()
		defer this.rwmu.Unlock()
		tpl = template.New(name)
		if this.handler != nil {
			this.handler(tpl)
		}
		tpl, err := tpl.Parse(string(html))
		if err != nil {
			if line, e, ok := isParseError(err); ok {
				scanner := bufio.NewScanner(bytes.NewReader(html))
				for i := 0; i < line && scanner.Scan(); i++ {}
				if err := scanner.Err(); err != nil {
					return err
				}
				// Ignore debug page
				var pgs []Page
				for _, p := range ps {
					if !p.Debug {
						pgs = append(pgs, p)
					}
				}
				return &ParseError {
					Line: scanner.Text(),
					Err: e,
					Pages: pgs,
				}
			}
			return err
		}
		err = tpl.Execute(w, data)
		if err != nil {
			return err
		}
		if !this.debug {
			this.tpls[name] = tpl
		}
	}
	return nil
}

var patten = regexp.MustCompile(`:(\d+):`)

func isParseError(e error) (line int, err string, ok bool) {
	// error: template: /home/zp/project/view/index.tpl:65: function "g" not defined
	errStr := e.Error()
	strs := patten.FindStringSubmatch(errStr)
	if len(strs) == 2 {
		// strs = [":65:" "65"]
		line, _ = strconv.Atoi(strs[1])
		start := strs[0]
		err = errStr[strings.Index(errStr, start) + len(start) + 1:]
		return line, err, true
	}
	return 0, "", false
}