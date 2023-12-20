package main

type FieldInfo struct {
	TypeName  string
	FieldName string
	Number    uint64
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
