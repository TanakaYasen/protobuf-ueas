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

var protoNative = map[protoreflect.Kind]typeMapper{
	protoreflect.BoolKind:   {"bool", "false", "DecodeBool()", "DecodeRepBool(%s_)", "EncodeBool(%d, %s_)", "EncodeRepBool(%d, %s_)"},
	protoreflect.Int32Kind:  {"int32_t", "0", "DecodeInt32()", "DecodeRepInt32(%s_)", "EncodeInt32(%d, %s_)", "EncodeRepInt32(%d, %s_)"},
	protoreflect.Sint32Kind: {"int32_t", "0", "DecodeSint32()", "DecodeRepSint32(%s_)", "EncodeSint32(%d, %s_)", "EncodeRepSint32(%d, %s_)"},
	protoreflect.Uint32Kind: {"uint32_t", "0", "DecodeUint32()", "DecodeRepUint32(%s_)", "EncodeUint32(%d, %s_)", "EncodeRepUint32(%d, %s_)"},
	protoreflect.Int64Kind:  {"int64_t", "0", "DecodeInt64()", "DecodeRepInt64(%s_)", "EncodeInt64(%d, %s_)", "EncodeRepInt64(%d, %s_)"},
	protoreflect.Sint64Kind: {"int64_t", "0", "DecodeSint64()", "DecodeRepSint64(%s_)", "EncodeSint64(%d, %s_)", "EncodeRepSint64(%d, %s_)"},
	protoreflect.Uint64Kind: {"uint64_t", "0", "DecodeUint64()", "DecodeRepUint64(%s_)", "EncodeUint64(%d, %s_)", "EncodeRepUint64(%d, %s_)"},

	protoreflect.Sfixed32Kind: {"int32_t", "0", "DecodeSfixed32()", "DecodeRepSfixed32(%s_)", "EncodeSfixed32(%d, %s_)", "EncodeRepSfixed32(%d, %s_)"},
	protoreflect.Fixed32Kind:  {"uint32_t", "0", "DecodeFixed32()", "DecodeRepFixed32(%s_)", "EncodeFixed32(%d, %s_)", "EncodeRepFixed32(%d, %s_)"},
	protoreflect.Sfixed64Kind: {"int64_t", "0", "DecodeSfixed64()", "DecodeRepSfixed64(%s_)", "EncodeSfixed64(%d, %s_)", "EncodeRepSfixed64(%d, %s_)"},
	protoreflect.Fixed64Kind:  {"uint64_t", "0", "DecodeFixed64()", "DecodeRepFixed64(%s_)", "EncodeFixed64(%d, %s_)", "EncodeRepFixed64(%d, %s_)"},

	protoreflect.FloatKind:  {"float", "0.f", "DecodeFloat()", "DecodeRepFloat(%s_)", "EncodeFloat(%d, %s_)", "EncodeRepFloat(%d, %s_)"},
	protoreflect.DoubleKind: {"double", "0.0", "DecodeDouble()", "DecodeRepDouble(%s_)", "EncodeDouble(%d, %s_)", "EncodeRepDouble(%d, %s_)"},
	protoreflect.StringKind: {"std::string", "", "DecodeString()", "", "EncodeString(%d, %s_)", ""},
	protoreflect.BytesKind:  {"std::vector<uint8_t>", "", "DecodeByte()", "", "EncodeBytes(%d, %s_)", ""},
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
			s.Printf("std::vector<%s> %s_;\n", n.cppType, fd.Name())
		} else {
			s.Printf("%s %s_;\n", n.cppType, fd.Name())
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
	sh.Printf("public:\n")

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

	scpp.Printf("std::string %s::Serialize() const\n", scpp.CppScopeLocate())
	scpp.Printf("{\n")
	scpp.Enter()
	scpp.Printf("WireEncoder encoder;\n")
	for _, field := range msg.Fields {
		if f, ok := protoNative[field.Desc.Kind()]; ok {
			if field.Desc.Cardinality() == protoreflect.Repeated {
				if f.encodeRepMethod == "" {
					scpp.Printf("\tfor (const auto &v_ : %s_) {\n", field.Desc.Name())
					s := fmt.Sprintf("\t\tencoder.%s;\n", f.encodeMethod)
					scpp.Printf(s, field.Desc.Number(), "v")
					scpp.Printf("\t}\n")
				} else {
					s := fmt.Sprintf("\tencoder.%s;\n", f.encodeRepMethod)
					scpp.Printf(s, field.Desc.Number(), field.Desc.Name())
				}
			} else {
				s := fmt.Sprintf("\tencoder.%s;\n", f.encodeMethod)
				scpp.Printf(s, field.Desc.Number(), field.Desc.Name())
			}
		} else if field.Desc.Kind() == protoreflect.EnumKind {
			//ename := scpp.DescopedName(string(field.Desc.Enum().FullName().Name()))
			scpp.Printf("\tencoder.EncodeEnum(%d, (uint64_t)%s_);\n", field.Desc.Number(), field.Desc.Name())
		} else if field.Desc.Kind() == protoreflect.MessageKind {
			if field.Desc.Cardinality() == protoreflect.Repeated {
				scpp.Printf("\tfor (const auto &v_ : %s_) {\n", field.Desc.Name())
				scpp.Printf("\tauto v = v_.Serialize() ; encoder.EncodeString(%d, v);\n", field.Desc.Number())
				scpp.Printf("\t}\n")
			} else {
				scpp.Printf("\t{ auto v = %s_.Serialize() ; encoder.EncodeString(%d, v);}\n", field.Desc.Name(), field.Desc.Number())
			}
		}
	}
	scpp.Printf("return encoder.Dump();\n")
	scpp.Leave()
	scpp.Printf("}\n")

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
			if field.Desc.Cardinality() == protoreflect.Repeated {
				if f.decodeRepMethod == "" {
					scpp.Printf("\tthis->%s_.push_back(decoder.%s);\n", field.Desc.Name(), f.decodeMethod)
				} else {
					s := fmt.Sprintf("\tdecoder.%s;\n", f.decodeRepMethod)
					log.Println(s)
					scpp.Printf(s, field.Desc.Name())
				}
			} else {
				scpp.Printf("\tthis->%s_ = decoder.%s;\n", field.Desc.Name(), f.decodeMethod)
			}
		} else if field.Desc.Kind() == protoreflect.EnumKind {
			ename := scpp.DescopedName(string(field.Desc.Enum().FullName().Name()))
			scpp.Printf("\tthis->%s_ = (%s)decoder.DecodeEnum();\n", field.Desc.Name(), ename)
		} else if field.Desc.Kind() == protoreflect.MessageKind {
			if field.Desc.Cardinality() == protoreflect.Repeated {
				scpp.Printf("{ %s v; v.Unserialize(decoder.DecodeSubmessage());\n",
					scpp.DescopedName(string(field.Desc.Message().FullName())))
				scpp.Printf("\tthis->%s_.push_back(v);}\n", field.Desc.Name())
			} else {
				scpp.Printf("\tthis->%s_.Unserialize(decoder.DecodeSubmessage());\n", field.Desc.Name())
			}
		}

		scpp.Printf("break;\n")
	}
	scpp.Printf("default:\n")
	scpp.Printf("\tdecoder.DecodeUnknown();\n")
	scpp.Printf("break;\n")
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
