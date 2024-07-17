package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"log"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
)

//go:embed service.hhtml
var hRpcStubCodeH string

//go:embed service.cpphtml
var hRpcStubCodeCpp string

type SourceFile struct {
	Package string
	*protogen.File
}

func genCppSource(g *protogen.GeneratedFile, file *protogen.File, temptext string) {
	buf := bytes.Buffer{}
	writer := bufio.NewWriter(&buf)

	funcMap := template.FuncMap{
		"LowerIdent":       lowerIdent,
		"UpperIdent":       upperIdent,
		"LowerString":      lowerString,
		"TrimRight":        trimRight,
		"TrimLeft":         trimLeft,
		"TrimString":       trimString,
		"IsEmpty":          isEmpty,
		"IsNotEmpty":       isNotEmpty,
		"PostPrefix":       postPrefix,
		"GetDirection":     getDirection,
		"GetProtoFilePath": getProtoFilePath,
		"GetProtoFileBase": getProtoFileBase,
	}
	templ, err := template.New("cppstub").Funcs(funcMap).Parse(temptext)
	if err != nil {
		log.Fatalln(err)
		return
	}
	err = templ.Execute(writer, file)
	if err != nil {
		log.Fatalln(err)
		return
	}
	writer.Flush()
	g.P(buf.String())
	g.P()
}

func generateCppRpc(gen *protogen.Plugin, file *protogen.File) {
	baseFilename := getRelativePath(file) //, cpp, c# is not like go,
	hFilename := baseFilename + ".arpc.h"
	ccFilename := baseFilename + ".arpc.cc"

	g := gen.NewGeneratedFile(hFilename, file.GoImportPath)
	genCppSource(g, file, hRpcStubCodeH)

	g = gen.NewGeneratedFile(ccFilename, file.GoImportPath)
	genCppSource(g, file, hRpcStubCodeCpp)
}
