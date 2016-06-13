package view

import (
	"bytes"
	"strings"
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

// Update default head tag or add some new tags, only registered tag and it's attribute can
// be merged.
func SetHeadTags(tag ...Tag) {
	for _, t := range tag {
		_append := true
		for idx, dt := range headTags {
			if dt.Name == t.Name {
				headTags[idx] = t
				_append = false
			}
		}
		if _append {
			headTags = append(headTags, t)
		}
	}
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
		head, _ := GetFirstSection(p, "head")
		heads[idx] = head

		// get all body
		body := GetFirstSectionWithTag(p, "body")
		bodies[idx] = body

		// get all script
		ss := NewSection(p[len(body) + len(head):], "script")
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
			section := NewSection(h, tag.Name)
			priority := 0
			for section.Next(func(content []byte, attr map[string]string) {
				priority++
				current := (idx << 8) + priority
				m := []byte("<" + tag.Name)
				for _, a := range tag.Attr {
					kv := mergeAttr(attr, a)
					m = append(m, kv...)
				}
				if tag.HasContent {
					m = append(m, byte('>'))
					m = append(m, content...)
					m = append(m, []byte("</" + tag.Name + ">")...)
				} else {
					m = append(m, []byte("/>")...)
				}
				aStr := string(m)
				// cover the same value
				if headCache[tag.Name] == nil {
					headCache[tag.Name] = map[string]int{aStr: current}
				} else {
					headCache[tag.Name][aStr] = current
				}
			}) {}
		}

	}

	// read the last title
	doc := pages[titleIndex]
	doc = doc[:bytes.Index(doc, []byte("</title>")) + 8]
	buffer := bytes.NewBuffer(nil)
	// write last "title" file like "<!DOCTYPE html>...</title>" into buffer
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

func NewSection(page []byte, tag string) *section {
	return &section{page: page, tag: tag, tagLen: len(tag)}
}

func GetFirstSection(page []byte, tag string) (content []byte, attr map[string]string) {
	s := NewSection(page, tag)
	s.Next(func(tagContent []byte, att map[string]string) {
		content = tagContent
		attr = att
	})
	return
}

func GetFirstSectionWithTag(page []byte, tag string) (tagContent []byte) {
	s := NewSection(page, tag)
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

	var start, end int
	a := TrimHtmlSpace(n)
	kvs := strings.Split(string(a), " ")
	attr = make(map[string]string, len(kvs))
	for _, kv := range kvs {
		kvSlice := strings.Split(kv, "=")
		if len(kvSlice) == 1 {
			attr[kvSlice[0]] = ""
		} else {
			v := kvSlice[1]
			start = strings.Index(v, `"`)
			if start == -1 {
				start = strings.Index(v, `'`)
				end = strings.LastIndex(v, `'`)
			} else {
				end = strings.LastIndex(v, `"`)
			}
			attr[kvSlice[0]] = v[start + 1:end]
		}
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