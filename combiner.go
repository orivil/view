// Copyright 2016 orivil Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package view provide a file combiner for merging layout files, and
// also provide a template container for cache the templates.
package view

import (
	"io/ioutil"
	"path/filepath"
	"regexp"
	"bytes"
)

type Combiner struct {
	dir          string
	ext          string
	sections     map[string][]byte
	layout       []byte
	firstContent []byte
}

func NewCombiner(dir, ext string) *Combiner {
	return &Combiner{
		dir: dir,
		ext: ext,
		sections: make(map[string][]byte, 1),
	}
}

// Combine combines section file into one full file.
func Combine(dir, file, exe string) ([]byte, error) {

	return NewCombiner(dir, exe).Combine(file)
}

func (s *Combiner) Combine(file string) ([]byte, error) {
	content, err := s.getFileContent([]byte(file))
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

func (s *Combiner) getFileContent(file []byte) ([]byte, error) {

	return ioutil.ReadFile(filepath.Join(s.dir, string(file) + s.ext))
}

var includePatten = regexp.MustCompile(`@include\(["']([\w\/\.\-\_]+)["']\)`)

func (s *Combiner) compileInclude(content []byte) ([]byte, error) {
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
func (s *Combiner) merge() ([]byte, error) {
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

var extendsPatten = regexp.MustCompile(`^\s*@extends\(["']([\w\/\.\-\_]+)["']\)`)

// read section extends layout file name and get the layout content
func (s *Combiner) readLayout(content []byte) error {
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

func (s *Combiner) findSections(content []byte) {
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
