package view

import (
	"io/ioutil"
	"path/filepath"
	"regexp"
	"bytes"
)

type Compiler struct {
	dir          string
	ext          string
	sections     map[string][]byte
	layout       []byte
	firstContent []byte
}

func NewCompiler(dir, ext string) *Compiler {
	return &Compiler{
		dir: dir,
		ext: ext,
		sections: make(map[string][]byte, 1),
	}
}

// Compile for compile section file, if the file didn't extends a layout file,
// just compile include file and return it
func (s *Compiler) Compile(name string) ([]byte, error) {
	content, err := s.getFileContent([]byte(name))
	if err != nil {
		return nil, err
	}
	err = s.readLayout(content)
	if err != nil {
		return nil, err
	}
	s.findSections(content)
	s.firstContent = content
	return s.merge()
}

func (s *Compiler) getFileContent(file []byte) ([]byte, error) {

	return ioutil.ReadFile(filepath.Join(s.dir, string(file) + s.ext))
}

var includePatten = regexp.MustCompile(`@include\(["']([\w\/\.]+)["']\)`)

func (s *Compiler) compileInclude(content []byte) ([]byte, error) {
	result := includePatten.FindAllSubmatch(content, -1)
	for _, r := range result {
		name := r[1]
		c, err := s.getFileContent(name)
		if err != nil {
			return nil, err
		}
		content = bytes.Replace(content, r[0], c, 1)
	}
	return content, nil
}

var yieldPatten = regexp.MustCompile(`@yield\(["']([\w]+)["']\)`)

// merge all files
func (s *Compiler) merge() ([]byte, error) {
	if s.layout == nil {
		return s.compileInclude(s.firstContent)
	} else {
		result := yieldPatten.FindAllSubmatch(s.layout, -1)
		for _, r := range result {
			name := r[1]
			replace := []byte{}
			if section, ok := s.sections[string(name)]; ok {
				replace = section
			}
			s.layout = bytes.Replace(s.layout, r[0], replace, -1)
		}
	}
	return s.compileInclude(s.layout)
}

var extendsPatten = regexp.MustCompile(`^\s*@extends\(["']([\w\/\.]+)["']\)`)

// read section extends layout file name and get the layout content
func (s *Compiler) readLayout(content []byte) error {
	result := extendsPatten.FindAllSubmatch(content, -1)
	if len(result) > 0 {
		name := result[0][1]
		c, err := s.getFileContent(name)
		if err != nil {
			return err
		}
		s.layout = c
	}
	return nil
}

var endsectionPatten = regexp.MustCompile(`@endsection\s*$`)
var sectionPatten = regexp.MustCompile(`@section\(["']([\w]+)["']\)([\s\S]+)@endsection`)
var prefixPatten = regexp.MustCompile(`^[\s\n]*`)
var suffixPatten = regexp.MustCompile(`[\s\n]*$`)

func (s *Compiler) findSections(content []byte) {
	// auto add close tag
	if !endsectionPatten.Match(content) {

		content = append(content, []byte("@endsection")...)
	}

	result := sectionPatten.FindAllSubmatch(content, -1)
	if len(result) > 0 {
		name := string(result[0][1])
		matched := result[0][2]
		matched = prefixPatten.ReplaceAll(matched, []byte{})
		matched = suffixPatten.ReplaceAll(matched, []byte{})

		// get first section
		index := bytes.Index(matched, []byte("@endsection"))
		if index == -1 {
			s.sections[name] = matched
		} else {
			s.sections[name] = matched[0: index]
			// find next section
			s.findSections(matched[index:])
		}
	}
}
