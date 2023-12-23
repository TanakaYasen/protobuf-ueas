package main

import (
	"log"
	"strings"
)

type FieldInfo struct {
	TypeName     string
	FieldName    string
	Number       uint64
	EncodeMethod string
	DecodeMethod string
}

type ClassDef struct {
	ClassName string
	TypeName  string
	Fields    []*FieldInfo
	Nested    []*ClassDef
}

type ParsedStruct struct {
	SourceFile        string
	PackageName       string
	ProtoName         string
	HIncludes         []string
	CppIncludes       []string
	BracketIncludings []string
	ClassDefinations  []*ClassDef
}

type typeMapper struct {
	cpptype         string
	decodeMethod    string
	decodeRepMethod string
	encodeMethod    string
	encodeRepMethod string
}

type scopeResolver struct {
	scope []string
}

func (s *scopeResolver) ScopeIn(name string) {
	s.scope = append(s.scope, name)
}
func (s *scopeResolver) ScopeOut() {
	s.scope = s.scope[:len(s.scope)-1]
}

func (s *scopeResolver) DescopedName(name string) string {
	var segs = strings.Split(name, ".")
	for i, sc := range s.scope {
		if segs[i] != sc {
			log.Printf("%d %s", i, sc)
			return strings.Join(segs[i:], "::")
		}
	}
	return strings.Join(segs[len(s.scope):], "::")
}
func (s *scopeResolver) CppScopeLocate() string {
	return strings.Join(s.scope[1:], "::")
}
