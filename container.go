package view

import (
	"html/template"
	"sync"
	"io"
	"path/filepath"
)

type Container struct {
	tpls map[string]*template.Template
	ext   string
	debug bool
	handler func(tpl *template.Template)
	rwmu  *sync.RWMutex
}

// NewContainer
//
// debug: if debug is true, compiler will always read file from disk
// otherwise it will cache the file
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

// Clear for clear all cache
func (this *Container) Clear() {

	this.tpls = make(map[string]*template.Template, 15)
}

func (this *Container) Display(w io.Writer, dir, file string, data interface{}) error {

	var name string = filepath.Join(dir, file)
	this.rwmu.RLock()
	tpl, ok := this.tpls[name]
	this.rwmu.RUnlock()
	if ok {
		return tpl.Execute(w, data)
	} else {
		c := NewCompiler(dir, this.ext)
		html, err := c.Compile(file)
		if err != nil {
			return err
		}
		this.rwmu.Lock()
		defer this.rwmu.Unlock()
		tpl = template.New(name)
		if this.handler != nil {
			this.handler(tpl)
		}
		tpl, err = tpl.Parse(string(html))
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
