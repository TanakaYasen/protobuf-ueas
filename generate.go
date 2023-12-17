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

type typeMapper struct {
	cpptype string
	method  string
}

var protoNative = map[protoreflect.Kind]typeMapper{
	protoreflect.BoolKind:   {"bool", "DecodeBool()"},
	protoreflect.Int32Kind:  {"int32_t", "DecodeInt32()"},
	protoreflect.Sint32Kind: {"int32_t", "DecodeSint32()"},
	protoreflect.Uint32Kind: {"uint32_t", "DecodeUint32()"},
	protoreflect.Int64Kind:  {"int64_t", "DecodeInt64()"},
	protoreflect.Sint64Kind: {"int64_t", "DecodeSint64()"},
	protoreflect.Uint64Kind: {"uint64_t", "DecodeUint64()"},

	protoreflect.Sfixed32Kind: {"int32_t", "DecodeSFixed32()"},
	protoreflect.Fixed32Kind:  {"uint32_t", "DecodeFixed32()"},
	protoreflect.Sfixed64Kind: {"int64_t", "DecodeSFixed64()"},
	protoreflect.Fixed64Kind:  {"uint64_t", "DecodeFixed64()"},

	protoreflect.FloatKind:  {"float", "DecodeFloat()"},
	protoreflect.DoubleKind: {"double", "DecodeDouble()"},
	protoreflect.StringKind: {"std::string", "DecodeString()"},
	protoreflect.BytesKind:  {"std::vector<uint8_t>", "DecodeByte()"},
}

var protoCustom = map[protoreflect.Kind]typeMapper{
	protoreflect.EnumKind:    {"", ""},
	protoreflect.MessageKind: {"", ""},
}

const headerInclues = `
#pragma once

#include <cstdint>
#include <string>
#include <string_view>
#include <vector>

`
const cppIncludes = `
#include "addressbook.h"
#include "../runtime/wire.h"

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
func (s *stream) CppScopeLocate() string {
	return strings.Join(s.scoper[1:], "::")
}

func dumpField(s *stream, fd protoreflect.FieldDescriptor) {
	var isRepeated = fd.Cardinality() == protoreflect.Repeated
	if n, ok := protoNative[fd.Kind()]; ok {
		if isRepeated {
			s.Printf("std::vector<%s> %s_;\n", n.cpptype, fd.Name())
		} else {
			s.Printf("%s %s_;\n", n.cpptype, fd.Name())
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

func dumpClass(sh *stream, scpp *stream, msg *protogen.Message) {
	sh.ScopeIn(string(msg.Desc.Name()))
	defer sh.ScopeOut()
	scpp.ScopeIn(string(msg.Desc.Name()))
	defer scpp.ScopeOut()

	sh.Printf("class %s {\n", msg.Desc.Name())
	sh.Printf("private:\n")

	sh.Enter()
	for _, enum := range msg.Enums {
		sh.Printf("enum %s {\n", string(enum.Desc.Name()))

		sh.Enter()
		for _, ev := range enum.Values {
			sh.Printf("%s = %d,\n", ev.Desc.Name(), ev.Desc.Number())
		}
		sh.Leave()
		sh.Printf("};\n")
	}
	sh.Leave()

	for _, nested := range msg.Messages {
		sh.Enter()
		dumpClass(sh, scpp, nested)
		sh.Leave()
	}
	sh.Enter()

	for _, field := range msg.Fields {
		dumpField(sh, field.Desc)
	}

	sh.Leave()
	sh.Printf("public:\n")
	sh.Enter()
	sh.Printf("std::string Serialize() const;\n")
	sh.Printf("bool Unserialize(std::string_view sv);\n")
	sh.Leave()
	sh.Printf("};\n")

	scpp.Printf("bool %s::Unserialize(std::string_view sv)\n", scpp.CppScopeLocate())
	scpp.Printf("{\n")
	scpp.Enter()
	scpp.Printf("uint64_t fn;\n")
	scpp.Printf("WireDecoder decoder((const uint8_t*)sv.data(), sv.length());\n")
	scpp.Printf("while ((fn = decoder.ReadTag()) && decoder.IsOk()) {\n")
	scpp.Enter()
	scpp.Printf("switch(fn) {\n")
	scpp.Enter()
	for _, field := range msg.Fields {
		scpp.Printf("case %d:\n", field.Desc.Number())
		if f, ok := protoNative[field.Desc.Kind()]; ok {
			scpp.Printf("\tthis->%s_ = decoder.%s;\n", field.Desc.Name(), f.method)
		} else if field.Desc.Kind() == protoreflect.EnumKind {
			ename := scpp.DescopedName(string(field.Desc.Enum().FullName().Name()))
			scpp.Printf("\tthis->%s_ = (%s)decoder.DecodeEnum();\n", field.Desc.Name(), ename)
		} else if field.Desc.Kind() == protoreflect.MessageKind {
			scpp.Printf("\tauto sv = decoder.DecodeSubmessage();\n")
			scpp.Printf("\tthis->%s_.Unserialize(sv);\n", field.Desc.Name())
		}

		scpp.Printf("break;\n")
	}
	scpp.Printf("default:\n")
	scpp.Printf("\tdecoder.DecodeUnknown();\n")
	scpp.Leave()
	scpp.Printf("}\n")
	scpp.Leave()
	scpp.Printf("}\n")
	scpp.Printf("return decoder.IsOk();\n")
	scpp.Leave()
	scpp.Printf("}\n")
}

func generateFile(gen *protogen.Plugin, file *protogen.File) {
	var baseName string = path.Base(file.Desc.Path())
	baseName = baseName[:len(baseName)-len(filepath.Ext(baseName))]
	outputCpp := "./generated/" + baseName + ".cpp"
	outputHeader := "./generated/" + baseName + ".h"

	log.Println(file.GeneratedFilenamePrefix+".md", file.GoImportPath)

	outputCppFile, err := os.OpenFile(outputCpp, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer outputCppFile.Close()

	outputHeaderFile, err := os.OpenFile(outputHeader, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer outputHeaderFile.Close()

	streamerHeader := &stream{outputHeaderFile, 0, nil}
	streamerCpp := &stream{outputCppFile, 0, nil}
	streamerHeader.Printf(headerInclues)
	streamerCpp.Printf(cppIncludes)

	streamerHeader.Printf("namespace %s {\n", *file.Proto.Package)
	streamerCpp.Printf("namespace %s {\n", *file.Proto.Package)
	streamerHeader.ScopeIn(*file.Proto.Package)
	streamerCpp.ScopeIn(*file.Proto.Package)

	for _, msg := range file.Messages {
		streamerCpp.Enter()
		dumpClass(streamerHeader, streamerCpp, msg)
		streamerCpp.Leave()
		streamerCpp.Printf("\n\n")
	}

	streamerCpp.ScopeOut()
	streamerHeader.ScopeOut()
	streamerCpp.Printf("}\n")
	streamerHeader.Printf("}\n")
}
