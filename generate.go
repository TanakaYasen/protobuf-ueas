package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var cppNative = map[protoreflect.Kind]string{
	protoreflect.BoolKind:   "bool",
	protoreflect.Int32Kind:  "int32_t",
	protoreflect.Sint32Kind: "int32_t",
	protoreflect.Uint32Kind: "uint32_t",
	protoreflect.Int64Kind:  "int64_t",
	protoreflect.Sint64Kind: "int64_t",
	protoreflect.Uint64Kind: "uint64_t",
	protoreflect.FloatKind:  "float",
	protoreflect.DoubleKind: "double",
	protoreflect.StringKind: "std::string",
	protoreflect.BytesKind:  "std::vector<uint8_t>",
}

const cppIncludes = `
#include <cstdint>
#include <string>
#include <string_view>
#include <vector>

`

type stream struct {
	f      *os.File
	indent int
	scoper []string
}

func (s *stream) Printf(format string, a ...any) {
	fmt.Fprintf(s.f, "%s", strings.Repeat("\t", s.indent))
	fmt.Fprintf(s.f, format, a...)
}

func (s *stream) Enter() {
	s.indent++
}

func (s *stream) Leave() {
	s.indent--
}

func (s *stream) ScopeIn(name string) {
	s.scoper = append(s.scoper, name)
}
func (s *stream) ScopeOut() {
	s.scoper = s.scoper[0 : len(s.scoper)-1]
}
func (s *stream) DescopedName(name string) string {
	var segs = strings.Split(name, ".")
	for i, sc := range s.scoper {
		if segs[i] != sc {
			log.Printf("%d %s", i, sc)
			return strings.Join(segs[i:], "::")
		}
	}
	return strings.Join(segs[len(s.scoper):], "::")
}

func dumpField(s *stream, fd protoreflect.FieldDescriptor) {
	var isRepeated = fd.Cardinality() == protoreflect.Repeated
	if cppType, ok := cppNative[fd.Kind()]; ok {
		if isRepeated {
			s.Printf("std::vector<%s> %s_;\n", cppType, fd.Name())
		} else {
			s.Printf("%s %s_;\n", cppType, fd.Name())
		}
	} else {
		if fd.Kind() == protoreflect.MessageKind {
			if isRepeated {
				s.Printf("std::vector<%s> %s_;\n",
					s.DescopedName(string(fd.Message().FullName())),
					fd.Name())
			} else {
				s.Printf("%s %s_;\n",
					s.DescopedName(string(fd.Message().FullName())),
					fd.Name())
			}
		} else if fd.Kind() == protoreflect.EnumKind {
			s.Printf("%s %s_;\n",
				s.DescopedName(string(fd.Enum().FullName())),
				fd.Name())
		} else {

		}
	}

}
func dumpClass(s *stream, msg *protogen.Message) {
	s.ScopeIn(string(msg.Desc.Name()))
	defer s.ScopeOut()

	s.Printf("class %s {\n", msg.Desc.Name())
	s.Printf("private:\n")

	s.Enter()
	for _, enum := range msg.Enums {
		s.Printf("enum %s {\n", string(enum.Desc.Name()))

		s.Enter()
		for _, ev := range enum.Values {
			s.Printf("%s = %d,\n", ev.Desc.Name(), ev.Desc.Number())
		}
		s.Leave()
		s.Printf("};\n")
	}
	s.Leave()

	for _, nested := range msg.Messages {
		s.Enter()
		dumpClass(s, nested)
		s.Leave()
	}
	s.Enter()

	for _, field := range msg.Fields {
		dumpField(s, field.Desc)
	}

	s.Leave()
	s.Printf("public:\n")
	s.Enter()
	s.Printf("std::string Serialize() const;\n")
	s.Printf("bool Unserialize(std::string_view sv);\n")
	s.Leave()
	s.Printf("};\n")
}

func generateFile(gen *protogen.Plugin, file *protogen.File) {
	var baseName string = path.Base(file.Desc.Path())
	baseName = baseName[:len(baseName)-len(filepath.Ext(baseName))]
	outputPath := "./generated/" + baseName + ".cpp"

	log.Println(file.GeneratedFilenamePrefix+".md", file.GoImportPath)

	outputFile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer outputFile.Close()

	streamer := &stream{outputFile, 0, nil}
	streamer.Printf(cppIncludes)

	streamer.Printf("namespace %s {\n", *file.Proto.Package)
	streamer.ScopeIn(*file.Proto.Package)
	for _, msg := range file.Messages {
		streamer.Enter()
		dumpClass(streamer, msg)
		streamer.Leave()
		streamer.Printf("\n\n")
	}
	streamer.ScopeOut()
	streamer.Printf("}\n")
}
