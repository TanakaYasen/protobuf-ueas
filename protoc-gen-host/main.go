package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"google.golang.org/protobuf/compiler/protogen"
)

// Wire Protocol
// https://protobuf.dev/programming-guides/encoding/

func main() {
	debug.SetTraceback("crash")

	logFile, err := os.OpenFile("./app.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	protogen.Options{}.Run(func(plugin *protogen.Plugin) error {
		log.Println("PARAMs:", plugin.Request.GetParameter())
		for _, file := range plugin.Files {
			if !file.Generate {
				continue
			}
			generateGoRpc(plugin, file)
		}
		return nil
	})
}
