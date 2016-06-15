package view_test

import (
	"testing"
	"reflect"
	"gopkg.in/orivil/view.v0"
	"bytes"
)

type t struct {
	page     []byte
	tag      string
	section  string
	attr     map[string]string
}

var sectionTestData = []t{
	{
		page: []byte(`<script src="/jquery.js">content</script>`),
		tag: "script",
		section: "content",
		attr: map[string]string{"src": "/jquery.js"},
	},

	{
		page: []byte(`<head><script src="/jquery.js">content</script></head>
		`),
		tag: "head",
		section: `<script src="/jquery.js">content</script>`,
		attr: nil,
	},

	{
		page: []byte(`<head><script src="/jquery.js">content</script></head>`),
		tag: "script",
		section: `content`,
		attr: map[string]string{"src": "/jquery.js"},
	},

	{
		page: []byte(`<link rel="stylesheet" href="/bootstrap.css">`),
		tag: "link",
		section: "",
		attr: map[string]string{
			"rel": "stylesheet",
			"href": "/bootstrap.css",
		},
	},
}

func TestSection(t *testing.T) {

	for _, d := range sectionTestData {
		section := view.NewSection(d.page, d.tag)
		for section.Next(func(tc []byte, attr map[string]string) {
			tcStr := string(tc)
			if tcStr != d.section {
				t.Error("got content:", tcStr, "expect content:", d.section)
			}
			if !reflect.DeepEqual(attr, d.attr) {
				t.Error("got attr:", attr, "expect attr:", d.attr)
			}
		}) {}
	}
}

var spaceHtml = []map[string]string{
	map[string]string{
		"input": "   html   ",
		"output": "html",
	},

	map[string]string{
		"input": `  <  head  >
    <  title  >   orivil web framework < /title >
    <  script src  =  "/js/local.js" >  script  < /script >
<  /head >  `,

		"output": `<head><title>orivil web framework</title><script src="/js/local.js">script</script></head>`,
	},
}

func TestTrimHtmlSpace(t *testing.T) {
	for _, html := range spaceHtml {
		result := string(view.TrimHtmlSpace([]byte(html["input"])))
		if html["output"] != result {
			t.Error("got:", result, "need:", html["output"])
		}
	}
}

var mergeData = [][]byte{
	[]byte(`<!DOCTYPE html>
<html>
<head>
    <title>title-1</title>
    <link rel="stylesheet" href="/local-1.css"/>
    <script src="/local-1.js"></script>
</head>
<body class="class-1">
</body>
<script>
    script-1
</script>
</html>`),

	[]byte(`<!DOCTYPE html>
<html>
<head>
    <title>title-2</title>
    <link rel="stylesheet" href="/local-2.css"/>
    <script src="/local-2.js"></script>
</head>
<body class="class-2">
</body>
<script>
    script-2
</script>
</html>`),

	[]byte(`<!DOCTYPE html>
<html>
<head>
    <title>title-3</title>
    <link rel="stylesheet" href="/local-3.css"/>
    <script src="/local-3.js"></script>
</head>
<body class="class-3">
</body>
<script>
    script-3
</script>
</html>`),

	[]byte(`<!DOCTYPE html>
<html>
<body class="class-4">
</body>
<script>
    script-4
</script>
</html>`),

	[]byte(`<!DOCTYPE html>
<html>
<script>
    script-5
</script>
</html>`),

	[]byte(`<!DOCTYPE html>
<html>
<body class="class-6">
</body>
</html>`),

	[]byte(`<!DOCTYPE html>
<html>
<head>
    <title>title-3</title>
    <link rel="stylesheet" href="/local-1.css"/>
    <link rel="stylesheet" href="/local-2.css"/>
    <link rel="stylesheet" href="/local-3.css"/>
    <script src="/local-1.js"></script>
    <script src="/local-2.js"></script>
    <script src="/local-3.js"></script>
</head>
<body>
<div class="class-1">
</div>
<div class="class-2">
</div>
<div class="class-3">
</div>
<div class="class-4">
</div>
<div class="class-6">
</div>
</body>
<script>
    script-1
</script>
<script>
    script-2
</script>
<script>
    script-3
</script>
<script>
    script-4
</script>
<script>
    script-5
</script>
</html>`),
}

func TestMergeHtml(t *testing.T) {
	result := view.MergeHtml(mergeData[0:6])
	if !bytes.Equal(result, mergeData[6]) {
		t.Error("got:", string(result), "expect:", mergeData[6])
	}
}

func BenchmarkMergeHtml(b *testing.B) {
	for i := 0; i < b.N; i++ {
		view.MergeHtml(mergeData[0:6])
	}
}