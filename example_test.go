package view_test
import (
	"github.com/orivil/view"
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

	// the test was failed, but result just like this

	// Output:
	// <layout>
	//     <bar></bar>
	//
	//     <index>
	//         {{.}}
	//     </index>
	//
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

	// the test was failed, but result just like this

	// Output:
	// <layout>
	//     <bar></bar>
	//
	//     <index>
	//         hello world!
	//     </index>
	//
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

	// 3. display
	data := 28
	err := container.Display(os.Stdout, dir, "functest", data)
	if err != nil {
		log.Fatal(err)
	}

	// the right result is 46, this "460" is for see the the right result

	// Output:
	// 460
}