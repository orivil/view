package view

import (
	"html/template"
	"sync"
	"io"
	"path/filepath"
	"bytes"
)

type Container struct {
	tpls map[string]*template.Template
	ext   string
	debug bool
	handler func(tpl *template.Template)
	rwmu  *sync.RWMutex
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
	Dir string
	File string
}

func NewPage(dir, file string) Page {
	return Page{Dir: dir, File: file}
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
		var pages = make([][]byte, len(ps))
		var html []byte
		for idx, s := range ps {
			html, err := Combine(string(s.Dir), string(s.File), this.ext)
			if err != nil {
				return err
			}
			pages[idx] = html
		}
		if len(pages) > 1 {
			html = MergeHtml(pages)
		} else {
			html = pages[0]
		}
		this.rwmu.Lock()
		defer this.rwmu.Unlock()
		tpl = template.New(name)
		if this.handler != nil {
			this.handler(tpl)
		}
		tpl, err := tpl.Parse(string(html))
		if err != nil {
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
