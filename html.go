package view

import (
	"bytes"
	"unicode"
	"gopkg.in/orivil/sorter.v0"
)

type Tag struct {
	Name       string
	HasContent bool
	Attr       []string
}

// default head tags
var headTags = []Tag{
	Tag{
		Name: "meta",
		HasContent: false,
		Attr: []string{
			"name",
			"http-equiv",
			"content",
			"charset",
		},
	},

	Tag{
		Name: "link",
		HasContent: false,
		Attr: []string{
			"rel",
			"href",
			"media",
			"type",
			"sizes",
		},
	},

	Tag{
		Name: "script",
		HasContent: true,
		Attr: []string{
			"type",
			"async",
			"src",
			"charset",
			"defer",
		},
	},

	Tag{
		Name: "style",
		HasContent: true,
		Attr: []string{
			"media",
			"type",
		},
	},
}

// SetDefaultHeadTags replaces default head tags and their attributes.
// Only registered tags and their attributes can be merged.
func SetDefaultHeadTags(tags []Tag) {

	headTags = tags
}

// MergeHtml merges multiple html pages into one single page.
func MergeHtml(pages [][]byte) []byte {
	var (
		heads = make([][]byte, len(pages))
		bodies = make([][]byte, len(pages))
		scripts = make([][][]byte, len(pages))
		headCache = make(map[string]map[string]int)
		titleIndex int
	)

	// get all sections
	for idx, p := range pages {
		// get all heads
		head, _ := getFirstSection(p, "head")
		heads[idx] = head

		// get all body
		body := getFirstSectionWithTag(p, "body")
		bodies[idx] = body

		// get all script
		ss := newSection(p[bytes.LastIndex(p, []byte("</body>")):], "script")
		for ss.NextWithTag(func(tagContent []byte) {
			scripts[idx] = append(scripts[idx], tagContent)
		}) {}
	}

	// merge all heads
	for idx, h := range heads {
		if bytes.Index(h, []byte("<title")) != -1 {
			// mark the last title
			titleIndex = idx
		}
		// format "<head>...</head>"
		for _, tag := range headTags {
			section := newSection(h, tag.Name)
			priority := 0
			for section.Next(func(content []byte, attr map[string]string) {
				priority++
				current := (idx << 8) + priority
				buf := bytes.NewBuffer(nil)
				buf.WriteString("<")
				buf.WriteString(tag.Name)
				for _, a := range tag.Attr {
					kv := mergeAttr(attr, a)
					buf.Write(kv)
				}
				if tag.HasContent {
					buf.WriteRune('>')
					buf.Write(content)
					buf.WriteString("</")
					buf.WriteString(tag.Name)
					buf.WriteRune('>')
				} else {
					buf.WriteString("/>")
				}
				aStr := buf.String()
				// cover the same value
				if headCache[tag.Name] == nil {
					headCache[tag.Name] = map[string]int{aStr: current}
				} else {
					headCache[tag.Name][aStr] = current
				}
			}) {}
		}
	}

	buffer := bytes.NewBuffer(nil)
	// read the last title
	doc := pages[titleIndex]
	part := doc[:bytes.Index(doc, []byte("<head>")) + 6]
	// write start file like "<!DOCTYPE html><head>" into buffer
	buffer.Write(part)
	buffer.WriteString("\n    ")
	doc = doc[bytes.Index(doc, []byte("<title>")): bytes.Index(doc, []byte("</title>")) + 8]
	// write the last title "<title>...</title>" into buffer
	buffer.Write(doc)

	// write all head tags into buffer
	for _, tag := range headTags {
		if ms, ok := headCache[tag.Name]; ok {
			sms := sorter.NewPrioritySorter(ms).Sort()
			for _, val := range sms {
				buffer.WriteString("\n    ")
				buffer.WriteString(val)
			}
		}
	}

	buffer.WriteString("\n</head>\n<body>")

	// write all body into buffer
	for _, body := range bodies {
		if body != nil {
			buffer.WriteString("\n<div")
			buffer.Write(body[5:len(body) - 5]) // turn "<body" ... "body>" to "<div" ... "div>"
			buffer.WriteString("div>")
		}
	}
	buffer.WriteString("\n</body>")

	// write all script tags which after body
	for _, ss := range scripts {
		for _, script := range ss {
			if script != nil {
				buffer.WriteString("\n")
				buffer.Write(script)
			}
		}
	}
	// close html tag
	buffer.WriteString("\n</html>")
	return buffer.Bytes()
}

func mergeAttr(attr map[string]string, key string) (kv []byte) {
	if v, ok := attr[key]; ok {
		kv = append([]byte(` ` + key + `="`), v...)
		kv = append(kv, []byte(`"`)...)
	}
	return
}

type section struct {
	page   []byte
	tag    string
	tagLen int
}

func newSection(page []byte, tag string) *section {
	return &section{page: page, tag: tag, tagLen: len(tag)}
}

func getFirstSection(page []byte, tag string) (content []byte, attr map[string]string) {
	s := newSection(page, tag)
	s.Next(func(tagContent []byte, att map[string]string) {
		content = tagContent
		attr = att
	})
	return
}

func getFirstSectionWithTag(page []byte, tag string) (tagContent []byte) {
	s := newSection(page, tag)
	s.NextWithTag(func(tag []byte) {
		tagContent = tag
	})
	return
}

func (s *section) NextWithTag(success func(tag []byte)) bool {
	start := bytes.Index(s.page, []byte("<" + s.tag))
	if start != -1 {
		end := bytes.Index(s.page, []byte("</" + s.tag + ">")) + s.tagLen + 3
		success(s.page[start: end])
		s.page = s.page[end:]
		return true
	}
	return false
}

func (s *section) Next(success func(tagContent []byte, attr map[string]string)) bool {
	start := bytes.Index(s.page, []byte("<" + s.tag))
	var content []byte
	var attr map[string]string
	if start != -1 {
		var nextStart int
		startContent := s.page[start:]
		close := bytes.Index(startContent, []byte(">"))
		if close > 1 {
			if startContent[close-1] == '/' {
				close--
			}
		}
		attrStart := 1 + s.tagLen // "<tag "
		if attrStart < close {
			attr = getAttr(startContent[attrStart:close])
			nextStart = start + close + 1
		}
		if close > 0 {
			end := bytes.Index(startContent, []byte("</" + s.tag + ">"))
			if end > close {
				content = startContent[close + 1: end]
				nextStart = start + end + 3 + s.tagLen
			}
		}
		s.page = s.page[nextStart:]
		success(content, attr)
		return true
	} else {
		return false
	}
}

func getAttr(n []byte) (attr map[string]string) {

	a := TrimHtmlSpace(n)

	attr = make(map[string]string)
	left := a
	var key, value []rune
	for {
		key, left = readKey(left)
		if len(left) <= 1 {
			attr[string(key)] = ""
			return
		}
		value, left = readValue(left)
		attr[string(key)] = string(value)
		if len(left) <= 1 {
			return
		}
	}

	return
}

func readKey(s []rune) (key []rune, left []rune) {
	if s[0] == ' ' { // cut the first space
		s = s[1:]
	}
	for _, b := range s {
		if b == '=' || b == ' ' {
			left = s[len(key):]
			break
		} else {
			key = append(key, rune(b))
		}
	}
	return
}

func readValue(s []rune) (value []rune, left []rune) {
	if len(s)>2 && string(s[0:2]) == `="` {
		s = s[2:]
		for _, b := range s {
			if b == '"' {
				left = s[len(value) + 1:] // cut the last `"`
				break
			} else {
				value = append(value, rune(b))
			}
		}
	} else {
		left = s
	}
	return
}

var skipSpaceByPrev = func(prev rune) bool {
	return prev == '>' || prev == '<' || prev == '=' || prev == ';'
}

var skipSpaceByNext = func(next rune) bool {
	return next == '>' || next == '=' || next == ' ' || next == '<' || next == ';'
}

func TrimHtmlSpace(html []byte) []rune {
	html = bytes.TrimLeftFunc(html, func(r rune) bool {
		return unicode.IsSpace(r)
	})
	html = bytes.TrimRightFunc(html, func(r rune) bool {
		return unicode.IsSpace(r)
	})

	var rs []rune
	var hs = []rune(string(html))
	var n = len(hs) - 1 // skip the last rune
	if n < 3 {
		return hs
	}

	rs = append(rs, hs[0]) // append fist rune
	var prev = hs[0]
	for i := 1; i < n; i++ {
		r := hs[i]
		if unicode.IsSpace(r) {
			if !skipSpaceByPrev(prev) && !skipSpaceByNext(hs[i + 1]) {
				rs = append(rs, r)
			}
		} else {
			prev = r
			rs = append(rs, r)
		}
	}
	rs = append(rs, hs[n]) // append last rune
	return rs
}