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
		{{.TypeName}} {{.FieldName}}{{if .DefaultValue}} = {{.DefaultValue}}{{end}};
		{{end}}

		bool _Valid = true;
		ustring Serialize() const;
		bool Unserialize(const uint8*, size_t);
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

#pragma region "{{.ClassName}}"

	ustring {{.ClassName}}::Serialize() const {
{{if gt (len .Fields) 0}}
		WireEncoder encoder;
		{{range .Fields}}{{if .EncodeCode}}{{.EncodeCode}}{{else}}encoder.{{.EncodeMethod}}({{.Number}}, {{.FieldName}}){{end}};
		{{end}}
		return encoder.Dump();
{{else}}
		return ustring{};
{{end}}
	}
	bool {{.ClassName}}::Unserialize(const uint8* data, size_t len) {
{{if gt (len .Fields) 0}}
		uint64 fn = 0;
		WireDecoder decoder(data, len);
		while ((fn = decoder.ReadTag()) && decoder.IsOk()) {
			switch(fn) { {{range .Fields}}
			case {{.Number}}:
				{{if .DecodeCode}}{{.DecodeCode}}{{else}}{{.FieldName}} = {{.DecodeMethod}}();{{end}}
				break;{{end}}
			default:
				break;
			}
		}
		return decoder.Ok();
{{else}}
		return true;
{{end}}
	}
#pragma endregion //"{{.ClassName}}"
{{end}}
}
`

var ueasProtocolHeaderTempl = `
// this file is generated by tool
// DONT EDIT IT MANUALLY
// SOURCE: [[{{.SourceFile}}]]


`
var ueasProtocolCppTempl = `
// this file is generated by tool
// DONT EDIT IT MANUALLY
// SOURCE: [[{{.SourceFile}}]]


`

var ueasprotoNative = map[protoreflect.Kind]typeMapper{
	protoreflect.BoolKind:   {"bool", "false", "DecodeBool", "DecodeRepBool", "EncodeBool", "EncodeRepBool"},
	protoreflect.Int32Kind:  {"int32", "0", "DecodeInt32", "DecodeRepInt32", "EncodeInt32", "EncodeRepInt32"},
	protoreflect.Sint32Kind: {"int32", "0", "DecodeSint32", "DecodeRepSint32", "EncodeSint32", "EncodeRepSint32"},
	protoreflect.Uint32Kind: {"uint32", "0", "DecodeUint32", "DecodeRepUint32", "EncodeUint32", "EncodeRepUint32"},
	protoreflect.Int64Kind:  {"int64", "0", "DecodeInt64", "DecodeRepInt64", "EncodeInt64", "EncodeRepInt64"},
	protoreflect.Sint64Kind: {"int64", "0", "DecodeSint64", "DecodeRepSint64", "EncodeSint64", "EncodeRepSint64"},
	protoreflect.Uint64Kind: {"uint64", "0", "DecodeUint64", "DecodeRepUint64", "EncodeUint64", "EncodeRepUint64"},

	protoreflect.Sfixed32Kind: {"int32", "0", "DecodeSfixed32", "DecodeRepSfixed32", "EncodeSfixed32", "EncodeRepSfixed32"},
	protoreflect.Fixed32Kind:  {"uint32", "0", "DecodeFixed32", "DecodeRepFixed32", "EncodeFixed32", "EncodeRepFixed32"},
	protoreflect.Sfixed64Kind: {"int64", "0", "DecodeSfixed64", "DecodeRepSfixed64", "EncodeSfixed64", "EncodeRepSfixed64"},
	protoreflect.Fixed64Kind:  {"uint64", "0", "DecodeFixed64", "DecodeRepFixed64", "EncodeFixed64", "EncodeRepFixed64"},

	protoreflect.FloatKind:  {"float", "0.f", "DecodeFloat", "DecodeRepFloat", "EncodeFloat", "EncodeRepFloat"},
	protoreflect.DoubleKind: {"double", "0.0", "DecodeDouble", "DecodeRepDouble", "EncodeDouble", "EncodeRepDouble"},
	protoreflect.StringKind: {"FBinary", "", "DecodeString", "", "EncodeString", ""},
	protoreflect.BytesKind:  {"TArray<uint8>", "", "DecodeByte", "", "EncodeBytes", ""},
}

func parseMessageField(classDef *ClassDef, fd protoreflect.FieldDescriptor) {
	var isRepeated = fd.Cardinality() == protoreflect.Repeated
	var formater string = "%s"
	if isRepeated {
		formater = "TArray<%s>"
	}
	fieldInfo := &FieldInfo{
		FieldName: string(fd.Name()),
		Number:    uint64(fd.Number()),
	}

	if n, ok := ueasprotoNative[fd.Kind()]; ok {
		fieldInfo.TypeName = fmt.Sprintf(formater, n.cppType)
		if isRepeated {
			fieldInfo.EncodeMethod = n.encodeRepMethod
			fieldInfo.DecodeMethod = n.decodeRepMethod
		} else {
			fieldInfo.EncodeMethod = n.encodeMethod
			fieldInfo.DecodeMethod = n.decodeMethod
			fieldInfo.DefaultValue = n.zeroValue
		}
	} else {
		if fd.Kind() == protoreflect.MessageKind {
			fieldInfo.TypeName = fmt.Sprintf(formater, string(fd.Message().FullName()))

			if isRepeated {
				fieldInfo.EncodeCode = fmt.Sprintf("for (const auto &v: %s_){ encoder.EncodeSubmessage(%d, v); }",
					fieldInfo.FieldName, fieldInfo.Number)
				fieldInfo.DecodeCode = fmt.Sprintf("{ decoder.DecodeRepSubmessage(%s_); }",
					fieldInfo.FieldName)
			} else {
				fieldInfo.EncodeCode = fmt.Sprintf("encoder.EncodeSubmessage(%d, %s_);",
					fieldInfo.Number, fieldInfo.FieldName)
				fieldInfo.DecodeCode = fmt.Sprintf("{auto v = decoder.DecodeSubmessage(); %s_.Unserialize((const uint8*)v.Data(), v.Len()); };",
					fieldInfo.FieldName)
			}

		} else if fd.Kind() == protoreflect.EnumKind {

			fieldInfo.TypeName = fmt.Sprintf(formater, string(fd.Enum().FullName()))

		} else {

		}
	}

	classDef.Fields = append(classDef.Fields, fieldInfo)
}

func parseMessageClass(out *ParsedStruct, msg *protogen.Message) {
	var newClass *ClassDef = new(ClassDef)
	newClass.ClassName = string(msg.Desc.FullName().Name())

	newClass.Nested = make([]*ClassDef, 0)
	for _, subMessage := range msg.Messages {
		parseMessageClass(out, subMessage)
		//log.Println(subMessage)
	}

	for _, field := range msg.Fields {
		parseMessageField(newClass, field.Desc)
	}

	out.ClassDefinations = append(out.ClassDefinations, newClass)
}

func generateUeas(gen *protogen.Plugin, file *protogen.File) {
	pathStr := file.Desc.Path()
	var baseName string = path.Base(file.Desc.Path())
	baseName = baseName[:len(baseName)-len(filepath.Ext(baseName))]

	OutputDir := path.Join("Protobuf", "ProtobufUEAS", "Source", "Generated")
	outputHeader := filepath.Join(OutputDir, baseName+".h")
	outputCpp := filepath.Join(OutputDir, baseName+".cpp")

	pStruct := ParsedStruct{
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
			"uewire.h",
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
		parseMessageClass(&pStruct, msg)
	}

	templ, err := template.New("ueash").Parse(ueasHeaderTempl)
	if err != nil {
		log.Fatalln(err)
		return
	}
	templ.Execute(outputHeaderFile, pStruct)

	templ, err = templ.New("ueascpp").Parse(ueasCppTempl)
	if err != nil {
		log.Fatalln(err)
		return
	}
	templ.Execute(outputCppFile, pStruct)

}
