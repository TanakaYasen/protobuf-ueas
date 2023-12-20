package main

import (
	"os"
	"text/template"
)

const hCode = `
// This file is generated by tool, DONT EDIT IT
// Source: [[{{.SourceFile}}]]

{{range .BracketIncludings}}#include <{{.}}>
{{end}}

{{ range .ClassDefinations}}

class {{.ClassName}} {
public:

private:

};

{{end}}
`

func buildCppCode() {
	//var cs = cppStruct{}

	templ, err := template.New("test").Parse(hCode)
	if err != nil {
		return
	}

	var cpp = ParsedStruct{
		SourceFile:        "game.proto",
		BracketIncludings: []string{"cstdio", "cstdint"},
	}
	templ.Execute(os.Stdout, cpp)
}
