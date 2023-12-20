package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var ueasHeaderTempl string = `
// this file is generated by tool
// DONT EDIT IT MANUALLY
// SOURCE [[.SourceFile]]

# pragma once

{{range .HIncludes}}#include <{{.}}>
{{end}}

namespace {{.PackageName}} {
		
	{{range .ClassDefinations}}
	USTRUCT(BlueprintType)
	struct FSproto_{{ .ClassName }} {
	public:
		{{range .Fields}}
		UPROPERTY(BlueprintReadWrite)
		{{.TypeName}} {{.FieldName}};
		{{end}}
	};
	{{end}}
}

`

var ueasCppTempl string = `
// this file is generated by tool
// DONT EDIT IT MANUALLY
// SOURCE: [[{{.SourceFile}}]]

{{range .CppIncludes}}#include "{{.}}"
{{end}}

namespace {{.PackageName}} {
{{range .ClassDefinations}}

	void {{.ClassName}}::Serialize() const {

	}
	bool {{.ClassName}}::Unserialize() {
		uint64 fn;
		WireDecoder decoder((const uint8_t*)sv.data(), sv.length());
		while ((fn = decoder.ReadTag()) && decoder.IsOk()) {
			switch(fn) {
			{{range .Fields}}
			case {{.Number}}:
				{{.FieldName}} = decoder.Decode();
				break;{{end}}
			default:
				break;
			}
		}
	}

{{end}}
}
`

var xprotoNative = map[protoreflect.Kind]typeMapper{
	protoreflect.BoolKind:   {"bool", "DecodeBool()", "DecodeRepBool(%s_)", "EncodeBool(%d, %s_)", "EncodeRepBool(%d, %s_)"},
	protoreflect.Int32Kind:  {"int32", "DecodeInt32()", "DecodeRepInt32(%s_)", "EncodeInt32(%d, %s_)", "EncodeRepInt32(%d, %s_)"},
	protoreflect.Sint32Kind: {"int32", "DecodeSint32()", "DecodeRepSint32(%s_)", "EncodeSint32(%d, %s_)", "EncodeRepSint32(%d, %s_)"},
	protoreflect.Uint32Kind: {"uint32", "DecodeUint32()", "DecodeRepUint32(%s_)", "EncodeUint32(%d, %s_)", "EncodeRepUint32(%d, %s_)"},
	protoreflect.Int64Kind:  {"int64", "DecodeInt64()", "DecodeRepInt64(%s_)", "EncodeInt64(%d, %s_)", "EncodeRepInt64(%d, %s_)"},
	protoreflect.Sint64Kind: {"int64", "DecodeSint64()", "DecodeRepSint64(%s_)", "EncodeSint64(%d, %s_)", "EncodeRepSint64(%d, %s_)"},
	protoreflect.Uint64Kind: {"uint64", "DecodeUint64()", "DecodeRepUint64(%s_)", "EncodeUint64(%d, %s_)", "EncodeRepUint64(%d, %s_)"},

	protoreflect.Sfixed32Kind: {"int32", "DecodeSfixed32()", "DecodeRepSfixed32(%s_)", "EncodeSfixed32(%d, %s_)", "EncodeRepSfixed32(%d, %s_)"},
	protoreflect.Fixed32Kind:  {"uint32", "DecodeFixed32()", "DecodeRepFixed32(%s_)", "EncodeFixed32(%d, %s_)", "EncodeRepFixed32(%d, %s_)"},
	protoreflect.Sfixed64Kind: {"int64", "DecodeSfixed64()", "DecodeRepSfixed64(%s_)", "EncodeSfixed64(%d, %s_)", "EncodeRepSfixed64(%d, %s_)"},
	protoreflect.Fixed64Kind:  {"uint64", "DecodeFixed64()", "DecodeRepFixed64(%s_)", "EncodeFixed64(%d, %s_)", "EncodeRepFixed64(%d, %s_)"},

	protoreflect.FloatKind:  {"float", "DecodeFloat()", "DecodeRepFloat(%s_)", "EncodeFloat(%d, %s_)", "EncodeRepFloat(%d, %s_)"},
	protoreflect.DoubleKind: {"double", "DecodeDouble()", "DecodeRepDouble(%s_)", "EncodeDouble(%d, %s_)", "EncodeRepDouble(%d, %s_)"},
	protoreflect.StringKind: {"FBinary", "DecodeString()", "", "EncodeString(%d, %s_)", ""},
	protoreflect.BytesKind:  {"TArray<uint8>", "DecodeByte()", "", "EncodeBytes(%d, %s_)", ""},
}

func dumpMessageField(classDef *ClassDef, fd protoreflect.FieldDescriptor) {
	var isRepeated = fd.Cardinality() == protoreflect.Repeated
	var formater string = "%s"
	if isRepeated {
		formater = "TArray<%s>"
	}
	fieldInfo := &FieldInfo{
		FieldName: string(fd.Name()),
		Number:    uint64(fd.Number()),
	}

	if n, ok := xprotoNative[fd.Kind()]; ok {
		fieldInfo.TypeName = fmt.Sprintf(formater, n.cpptype)
	} else {
		if fd.Kind() == protoreflect.MessageKind {

			fieldInfo.TypeName = fmt.Sprintf(formater, string(fd.Message().FullName()))

		} else if fd.Kind() == protoreflect.EnumKind {

			fieldInfo.TypeName = fmt.Sprintf(formater, string(fd.Enum().FullName()))

		} else {

		}
	}

	classDef.Fields = append(classDef.Fields, fieldInfo)
}

func addMessageClass(out *ParsedStruct, msg *protogen.Message) {
	var newClass *ClassDef = new(ClassDef)
	newClass.ClassName = string(msg.Desc.FullName().Name())
	for _, field := range msg.Fields {
		dumpMessageField(newClass, field.Desc)
	}

	out.ClassDefinations = append(out.ClassDefinations, newClass)
}

func generateUeas(gen *protogen.Plugin, file *protogen.File) {
	pathStr := file.Desc.Path()
	var baseName string = path.Base(file.Desc.Path())
	baseName = baseName[:len(baseName)-len(filepath.Ext(baseName))]
	outputCpp := "./generated/" + baseName + ".cpp"
	outputHeader := "./generated/" + baseName + ".h"

	pdata := ParsedStruct{
		SourceFile:  pathStr,
		ProtoName:   baseName,
		PackageName: *file.Proto.Package,
		HIncludes: []string{
			"CoreMinimal.h",
			"Container/TArray.h",
			"Container/TMap.h",
		},
		CppIncludes: []string{
			baseName + ".h",
		},
	}

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

	for _, msg := range file.Messages {
		addMessageClass(&pdata, msg)
	}

	templ, err := template.New("ueash").Parse(ueasHeaderTempl)
	if err != nil {
		log.Fatalln(err)
		return
	}
	templ.Execute(outputHeaderFile, pdata)

	templ, err = templ.New("ueascpp").Parse(ueasCppTempl)
	if err != nil {
		log.Fatalln(err)
		return
	}
	templ.Execute(outputCppFile, pdata)
}
