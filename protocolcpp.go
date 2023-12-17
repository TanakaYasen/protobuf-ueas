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

type FieldStruct struct {
	TypeName  string
	FieldName string
}

type ClassDef struct {
	ClassName string
	TypeName  string
	Fields    []FieldStruct
	Nested    []*ClassDef
}

type CppStruct struct {
	SourceFile        string
	BracketIncludings []string
	ClassDefinations  []ClassDef
}

func buildCppCode() {
	//var cs = cppStruct{}

	templ, err := template.New("test").Parse(hCode)
	if err != nil {
		return
	}

	var cpp = CppStruct{
		SourceFile:        "game.proto",
		BracketIncludings: []string{"cstdio", "cstdint"},
		ClassDefinations: []ClassDef{
			{"Zoo", "Game::Zoo", nil, nil},
			{"Animal", "Game::Animal", nil, nil},
		},
	}
	templ.Execute(os.Stdout, cpp)
}
