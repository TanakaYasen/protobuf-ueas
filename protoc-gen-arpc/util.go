package main

import (
	"bytes"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

func lowerIdent(s string) string {
	bts := []byte(s)
	for i, c := range bts {
		if i == 0 {
			if c >= 'A' && c <= 'Z' {
				c += 32
			} else {
				return s
			}
		} else {
			if c >= 'A' && c <= 'Z' {
				// 如果下一个字符还是大写,或者已经是末尾了,那就转为小写
				if i+1 < len(bts) {
					if bts[i+1] >= 'A' && bts[i+1] <= 'Z' {
						c += 32
					}
				} else {
					c += 32
				}
			} else {
				break
			}
		}
		bts[i] = c
	}
	return string(bts)
}

// server -> Server
func upperIdent(s string) string {
	bts := []byte(s)
	if len(bts) == 0 {
		return s
	}

	c := bts[0]
	if c >= 'a' && c <= 'z' {
		bts[0] = c - 32
	}
	return string(bts)
}

func lowerString(s string) string {
	return strings.ToLower(s)
}

func trimRight(s string) string {
	return string(bytes.TrimRight([]byte(s), " \t\r\n"))
}

func trimLeft(s string) string {
	return string(bytes.TrimLeft([]byte(s), " \t\r\n"))
}

func trimString(s string) string {
	return string(bytes.Trim([]byte(s), " \t\r\n"))
}

func isEmpty(s string) bool {
	return s == "Void"
}

func isNotEmpty(s string) bool {
	return s != "Void"
}

func postPrefix(s string) string {
	if s == "Void" {
		return "Send"
	}
	return "Call"
}

func getDirection(s protogen.Comments) string {
	if strings.Contains(strings.ToLower(s.String()), "s2c") {
		return "s2c"
	}
	if strings.Contains(strings.ToLower(s.String()), "c2s") {
		return "s2c"
	}
	return ""
}

func getProtoFilePath(file *protogen.File) string {
	return file.Desc.Path()
}

func getProtoFileBase(file *protogen.File) string {
	protoFilePath := file.Desc.Path()
	fileName := filepath.Base(protoFilePath)
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func getRelativePath(file *protogen.File) string {
	relativePath := file.Desc.Path()
	return relativePath[:len(relativePath)-len(filepath.Ext(relativePath))]
}
