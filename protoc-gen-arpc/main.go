package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

// Wire Protocol
// https://protobuf.dev/programming-guides/encoding/

var modeFuncs = map[string]func(gen *protogen.Plugin, file *protogen.File){
	"cpp": generateCppRpc,
	"go":  generateGoRpc,
}

func main() {
	debug.SetTraceback("crash")

	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-arpc %v\n", 123)
		return
	}

	logFile, err := os.OpenFile("./app.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	protogen.Options{}.Run(func(plugin *protogen.Plugin) error {
		params := plugin.Request.GetParameter()
		flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
		flag.CommandLine.Parse(strings.Split(params, ","))
		for _, file := range plugin.Files {
			if !file.Generate {
				continue
			}
			mf, ok := modeFuncs[params]
			if !ok {
				mf = generateCppRpc
			}
			mf(plugin, file)
		}
		return nil
	})
}
