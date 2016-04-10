package view_test
import (
	"gopkg.in/orivil/view.v0"
	"log"
	"fmt"
	"os"
	"html/template"
)


var dir = "./testdata"
var fileExt = ".blade.php"

func ExampleCompiler() {
	// 1. new compiler
	compiler := view.NewCompiler(dir, fileExt)

	// 2. compile files to one file
	file := "index"
	content, err := compiler.Compile(file)
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
	err := container.Display(os.Stdout, dir, file, data)
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

func ExampleSetTplHandle() {
	var debug = true

	// 1. new container
	container := view.NewContainer(debug, fileExt)

	// 2. handle new template
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
	err := container.Display(os.Stdout, dir, viewFile, data)
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// 46
}