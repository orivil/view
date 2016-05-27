package view_test
import (
	"gopkg.in/orivil/view.v0"
	"log"
	"fmt"
	"os"
	"html/template"
	"bytes"
	"io/ioutil"
)


var dir = "./testdata"
var fileExt = ".blade.php"

func ExampleCombiner() {
	// 1. new combiner
	combiner := view.NewCombiner(dir, fileExt)

	// 2. compile files to one file
	file := "index"
	content, err := combiner.Combine(file)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(content))

	// Output:
	// <layout>
	//     <bar></bar>
	//     <index>
	//         {{.}}
	//     </index>
	// </layout>
}

func ExampleContainer() {
	var debug = true

	// 1. new container
	container := view.NewContainer(debug, fileExt)

	// 2. display
	data := "hello world!"
	file := "index"
	err := container.Display(os.Stdout, data, view.NewPage(dir, file))
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// <layout>
	//     <bar></bar>
	//     <index>
	//         hello world!
	//     </index>
	// </layout>
}

func ExampleConfigureTemplate() {
	var debug = true

	// 1. new container
	container := view.NewContainer(debug, fileExt)

	// 2. handle new template. each time when container built a new template,
	// the callback will be run.
	container.SetTplHandle(func(newTpl *template.Template) {
		// add a function
		newTpl.Funcs(map[string]interface{}{

			"sum": func(i...int) int {
				sum := 0
				for _, a := range i {
					sum += a
				}
				return sum
			},
		})
	})

	data := 28
	viewFile := "functest" // "./testdata/functest.blade.php"
	// 3. display
	err := container.Display(os.Stdout, data, view.NewPage(dir, viewFile))
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// 46
}

func ExampleMergeHtml() {
	var debug = true

	// 1. new container
	container := view.NewContainer(debug, fileExt)

	// 2. display multiple files
	data := "hello world!"
	pages := []view.Page{
		{
			Dir: "./testdata/htmls/html-1",
			File: "index",
		},

		{
			Dir: "./testdata/htmls/html-2",
			File: "index",
		},

		{
			Dir: "./testdata/htmls/html-3",
			File: "index",
		},
	}

	// write to buffer
	buffer := bytes.NewBuffer(nil)
	err := container.Display(buffer, data, pages...)
	if err != nil {
		log.Fatal(err)
	}

	// the merged file was just like "./testdata/htmls/output/merged.html"
	result, _ := ioutil.ReadFile("./testdata/htmls/output/merged.html")
	isEqual := bytes.Equal(buffer.Bytes(), result)
	fmt.Println(isEqual)

	// Output:
	// true
}