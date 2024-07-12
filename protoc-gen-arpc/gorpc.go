package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"go/format"
	"log"

	"path/filepath"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
)

//go:embed service.gohtml
var hRpcStubCodeGo string

func genFmtSource(g *protogen.GeneratedFile, file *protogen.File, temptext string) {
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
	}
	templ, err := template.New("rpcstub").Funcs(funcMap).Parse(temptext)
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

	dat, err := format.Source(buf.Bytes())
	if err != nil {
		return
	}

	g.P(string(dat))
	g.P()
}

func generateGoRpc(gen *protogen.Plugin, file *protogen.File) {
	baseFilename := filepath.Base(file.GeneratedFilenamePrefix)
	filename := baseFilename + ".arpc.go"

	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	genFmtSource(g, file, hRpcStubCodeGo)
}
