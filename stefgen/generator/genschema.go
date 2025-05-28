package generator

import (
	"fmt"
	"strings"

	"github.com/splunk/stef/go/pkg/schema"
)

// genSchema is a STEF schema description in form that is useful for generation.
type genSchema struct {
	PackageName string
	Structs     map[string]*genStructDef
	Multimaps   map[string]*genMapDef
	Enums       map[string]*genEnumDef
}

func (s *genSchema) SchemaStr() string {
	str := ""
	for _, mp := range s.Multimaps {
		str += "multimap " + mp.Name + " {\n"
		str += "  Key " + mp.Key.Type.SchemaStr() + "\n"
		str += "  Value " + mp.Value.Type.SchemaStr() + "\n"
		str += "}\n\n"
	}
	for _, struc := range s.Structs {
		if struc.OneOf {
			str += "oneof"
		} else {
			str += "struct"
		}
		str += " " + struc.Name
		if struc.IsRoot {
			str += " root"
		}
		if struc.Dict != "" {
			str += " dict(" + struc.Dict + ")"
		}
		str += " {\n"
		for _, field := range struc.Fields {
			str += "  " + field.Name + " " + field.Type.SchemaStr()
			if field.Optional {
				str += " optional"
			}
			str += "\n"
		}
		str += "}\n\n"
	}

	//str += "main " + s.MainStruct + "\n"
	return str
}

type recursable interface {
	SetRecursive()
}

type genStructFieldDef struct {
	Name      string
	Type      genFieldTypeRef
	Optional  bool
	Recursive bool
}

func (a *genStructFieldDef) SetRecursive() {
	a.Recursive = true
}

type TypeFlags struct {
	// PassByPtr indicates that the value of this type is passed by pointer to Exported()
	// when it is a function parameter or when it is returned by a function.
	PassByPtr bool

	// StoreByPtr indicates that struct fields of the value of this type is stored as a
	// pointer to Exported(). If this is false that the fields are simply of Exported().
	StoreByPtr bool

	// TakePtr is true a pointer must be taken of the field to pass it as a parameter.
	TakePtr bool
}

type genFieldTypeRef interface {
	TypeName() string
	IsPrimitive() bool

	// Exported is the fully qualified exported (public) Go type.
	Exported() string

	// Storage is the underlying storage type.
	Storage() string

	// ToExported converts argument to exported type if necessary.
	ToExported(arg string) string

	// ToStorage converts argument to storage type if necessary.
	ToStorage(arg string) string

	EncoderType() string
	EqualFunc() string
	CompareFunc() string
	MustClone() bool
	DictName() string
	DictGoType() string
	IsDictPossible() bool
	SchemaStr() string

	Flags() TypeFlags
}

type genPrimitiveTypeRef struct {
	Lang Lang
	Type schema.PrimitiveFieldType
	Dict string
	Enum string
}

func (r *genPrimitiveTypeRef) IsPrimitive() bool {
	return true
}

func (r *genPrimitiveTypeRef) Flags() TypeFlags {
	return TypeFlags{
		PassByPtr:  false,
		StoreByPtr: false,
	}
}

func (r *genPrimitiveTypeRef) IsDictPossible() bool {
	switch r.Type {
	case schema.PrimitiveTypeString, schema.PrimitiveTypeBytes:
		return true
	default:
		return false
	}
}

func (r *genPrimitiveTypeRef) DictName() string {
	return r.Dict
}

// Storage returns the underlying type of fields.
func (r *genPrimitiveTypeRef) Storage() string {
	if r.Type == schema.PrimitiveTypeBytes {
		switch r.Lang {
		case LangGo:
			return "pkg.Bytes"
		case LangJava:
			return "Bytes"
		default:
			panic(fmt.Sprintf("unknown language %v", r.Lang))
		}
	}
	return r.TypeName()
}

// ToStorage converts the argument to the underlying type
// if the underlying type is different than the exported type.
// If the types are the same, no conversion is performed.
func (r *genPrimitiveTypeRef) ToStorage(arg string) string {
	if r.Enum != "" {
		return r.Storage() + "(" + arg + ")"
	}
	return arg
}

// ToExported converts the argument to the exported type
// if the underlying storage type is different than the exported type.
// If the types are the same, no conversion is performed.
func (r *genPrimitiveTypeRef) ToExported(arg string) string {
	if r.Enum != "" {
		return r.Enum + "(" + arg + ")"
	}
	return arg
}

// Exported returns the fully qualified Go type for this field type
// that will be used for exported setters/getters.
// The underlying storage type may be different (e.g. if it is an Enum)
// and is available via Storage().
func (r *genPrimitiveTypeRef) Exported() string {
	if r.Enum != "" {
		return r.Enum
	}
	if r.Type == schema.PrimitiveTypeBytes {
		return r.Storage()
	}
	return r.TypeName()
}

func (r *genPrimitiveTypeRef) TypeName() string {
	var s string

	switch r.Lang {
	case LangGo:
		switch r.Type {
		case schema.PrimitiveTypeInt64:
			s += "int64"
		case schema.PrimitiveTypeUint64:
			s += "uint64"
		case schema.PrimitiveTypeFloat64:
			s += "float64"
		case schema.PrimitiveTypeBool:
			s += "bool"
		case schema.PrimitiveTypeString:
			s += "string"
		case schema.PrimitiveTypeBytes:
			s += "Bytes"
		default:
			panic(fmt.Errorf("unimplemented field type %v", r.Type))
		}
	case LangJava:
		switch r.Type {
		case schema.PrimitiveTypeInt64:
			s += "long"
		case schema.PrimitiveTypeUint64:
			s += "long"
		case schema.PrimitiveTypeFloat64:
			s += "double"
		case schema.PrimitiveTypeBool:
			s += "boolean"
		case schema.PrimitiveTypeString:
			s += "StringValue"
		case schema.PrimitiveTypeBytes:
			s += "byte[]"
		default:
			panic(fmt.Errorf("unimplemented field type %v", r.Type))
		}
		if r.Enum != "" {
			return r.Enum
		}
	}

	return s
}

func (r *genPrimitiveTypeRef) EncoderType() string {
	var prefix string
	switch r.Lang {
	case LangGo:
		prefix = "encoders."
	case LangJava:
		prefix = ""
	default:
		panic(fmt.Sprintf("unknown language %v", r.Lang))
	}

	switch r.Type {
	case schema.PrimitiveTypeUint64:
		return prefix + "Uint64"
	case schema.PrimitiveTypeInt64:
		return prefix + "Int64"
	case schema.PrimitiveTypeFloat64:
		return prefix + "Float64"
	case schema.PrimitiveTypeBool:
		return prefix + "Bool"
	case schema.PrimitiveTypeString:
		return prefix + "String"
	case schema.PrimitiveTypeBytes:
		return prefix + "Bytes"
	default:
		panic(fmt.Sprintf("unknown type %v", r.Type))
	}
}

func (r *genPrimitiveTypeRef) DictGoType() string {
	var prefix string
	switch r.Lang {
	case LangGo:
		prefix = "encoders."
	case LangJava:
		prefix = ""
	default:
		panic(fmt.Sprintf("unknown language %v", r.Lang))
	}

	switch r.Type {
	case schema.PrimitiveTypeString:
		return prefix + "String"
	case schema.PrimitiveTypeBytes:
		return prefix + "Bytes"
	default:
		panic(fmt.Sprintf("type %v does not support dictionaries", r.Type))
	}
}

func (r *genPrimitiveTypeRef) pkgPrefix() string {
	var prefix string
	switch r.Lang {
	case LangGo:
		prefix = "pkg."
	case LangJava:
		prefix = "Types."
	default:
		panic(fmt.Sprintf("unknown language %v", r.Lang))
	}
	return prefix
}

func (r *genPrimitiveTypeRef) EqualFunc() string {
	if r.Lang == LangJava && r.Enum != "" {
		return r.Enum + ".equals"
	}

	prefix := r.pkgPrefix()
	switch r.Type {
	case schema.PrimitiveTypeUint64:
		return prefix + "Uint64Equal"
	case schema.PrimitiveTypeInt64:
		return prefix + "Int64Equal"
	case schema.PrimitiveTypeFloat64:
		return prefix + "Float64Equal"
	case schema.PrimitiveTypeBool:
		return prefix + "BoolEqual"
	case schema.PrimitiveTypeString:
		return prefix + "StringEqual"
	case schema.PrimitiveTypeBytes:
		return prefix + "BytesEqual"
	default:
		panic(fmt.Sprintf("unknown type %v", r.Type))
	}
}

func (r *genPrimitiveTypeRef) RandomFunc() string {
	prefix := r.pkgPrefix()
	switch r.Type {
	case schema.PrimitiveTypeUint64:
		return prefix + "Uint64Random"
	case schema.PrimitiveTypeInt64:
		return prefix + "Int64Random"
	case schema.PrimitiveTypeFloat64:
		return prefix + "Float64Random"
	case schema.PrimitiveTypeBool:
		return prefix + "BoolRandom"
	case schema.PrimitiveTypeString:
		return prefix + "StringRandom"
	case schema.PrimitiveTypeBytes:
		return prefix + "BytesRandom"
	default:
		panic(fmt.Sprintf("unknown type %v", r.Type))
	}
}

func (r *genPrimitiveTypeRef) CompareFunc() string {
	prefix := r.pkgPrefix()
	switch r.Type {
	case schema.PrimitiveTypeUint64:
		return prefix + "Uint64Compare"
	case schema.PrimitiveTypeInt64:
		return prefix + "Int64Compare"
	case schema.PrimitiveTypeFloat64:
		return prefix + "Float64Compare"
	case schema.PrimitiveTypeBool:
		return prefix + "BoolCompare"
	case schema.PrimitiveTypeString:
		return prefix + "StringCompare"
	case schema.PrimitiveTypeBytes:
		return prefix + "BytesCompare"
	default:
		panic(fmt.Sprintf("unknown type %v", r.Type))
	}
}

func (r *genPrimitiveTypeRef) MustClone() bool {
	return false
}

func (r *genPrimitiveTypeRef) SchemaStr() string {
	str := ""
	switch r.Type {
	case schema.PrimitiveTypeUint64:
		str = "uint64"
	case schema.PrimitiveTypeInt64:
		str = "int64"
	case schema.PrimitiveTypeFloat64:
		str = "float64"
	case schema.PrimitiveTypeBool:
		str = "bool"
	case schema.PrimitiveTypeString:
		str = "string"
	case schema.PrimitiveTypeBytes:
		str = "bytes"
	default:
		panic(fmt.Sprintf("unknown type %v", r.Type))
	}
	if r.Dict != "" {
		str += " dict(" + r.Dict + ")"
	}
	return str
}

type genStructDef struct {
	Name   string
	Dict   string
	Fields []*genStructFieldDef
	OneOf  bool
	IsRoot bool
}

type genMapFieldDef struct {
	Type      genFieldTypeRef
	Recursive bool
}

func (a *genMapFieldDef) SetRecursive() {
	a.Recursive = true
}

type genMapDef struct {
	Name  string
	Key   genMapFieldDef
	Value genMapFieldDef
}

type genStructTypeRef struct {
	Name string
	Def  *genStructDef
	Lang Lang
}

func (r *genStructTypeRef) IsPrimitive() bool {
	return false
}

func (r *genStructTypeRef) IsDictPossible() bool {
	return true
}

func (r *genStructTypeRef) DictName() string {
	return r.Def.Dict
}

func (r *genStructTypeRef) DictGoType() string {
	return r.DictName()
}

func (r *genStructTypeRef) Exported() string {
	return r.TypeName()
}

func (r *genStructTypeRef) Storage() string {
	return r.TypeName()
}

func (r *genStructTypeRef) ToExported(arg string) string {
	return arg
}

func (r *genStructTypeRef) ToStorage(arg string) string {
	return arg
}

func (r *genStructTypeRef) TypeName() string {
	return r.Name
}

func (r *genStructTypeRef) EncoderType() string {
	return r.Name
}

func (r *genStructTypeRef) EqualFunc() string {
	return r.Name + "Equal"
}

func (r *genStructTypeRef) CompareFunc() string {
	switch r.Lang {
	case LangGo:
		return "Cmp" + r.Name
	case LangJava:
		return r.Name + ".compare"
	default:
		panic(fmt.Sprintf("unknown language %v", r.Lang))
	}
}

func (r *genStructTypeRef) MustClone() bool {
	return true
}

func (r *genStructTypeRef) SchemaStr() string {
	return r.Name
}

func (r *genStructTypeRef) Flags() TypeFlags {
	return TypeFlags{
		PassByPtr:  true,
		StoreByPtr: r.DictName() != "",
		TakePtr:    r.DictName() == "",
	}
}

type genArrayTypeRef struct {
	ElemType genFieldTypeRef
	Lang     Lang
}

func (r *genArrayTypeRef) Flags() TypeFlags {
	return TypeFlags{
		PassByPtr:  true,
		StoreByPtr: false,
		TakePtr:    true,
	}
}

func (r *genArrayTypeRef) IsDictPossible() bool {
	return false
}

func (r *genArrayTypeRef) SchemaStr() string {
	return "[]" + r.ElemType.SchemaStr()
}

func (r *genArrayTypeRef) IsPrimitive() bool {
	return false
}

func (r *genArrayTypeRef) DictName() string {
	return ""
}

func (r *genArrayTypeRef) DictGoType() string {
	return ""
}

func (r *genArrayTypeRef) Exported() string {
	return r.TypeName()
}

func (r *genArrayTypeRef) Storage() string {
	return r.TypeName()
}

func (r *genArrayTypeRef) ToExported(arg string) string {
	return r.ToExported(arg)
}

func (r *genArrayTypeRef) ToStorage(arg string) string {
	return r.ToStorage(arg)
}

func (r *genArrayTypeRef) TypeName() string {
	str := r.ElemType.TypeName() + "Array"
	// Make sure the type name is exported.
	str = strings.ToUpper(str[:1]) + str[1:]
	return str
}

func (r *genArrayTypeRef) EncoderType() string {
	return r.TypeName()
}

func (r *genArrayTypeRef) EqualFunc() string {
	return r.TypeName() + "Equal"
}

func (r *genArrayTypeRef) CompareFunc() string {
	switch r.Lang {
	case LangGo:
		return "Cmp" + r.TypeName()
	case LangJava:
		return r.TypeName() + ".compare"
	default:
		panic(fmt.Sprintf("unknown language %v", r.Lang))
	}
}

func (r *genArrayTypeRef) MustClone() bool {
	return true
}

type genMultimapTypeRef struct {
	Name string
	Def  *genMapDef
}

func (r *genMultimapTypeRef) IsDictPossible() bool {
	return false
}

func (r *genMultimapTypeRef) IsPrimitive() bool {
	return false
}

func (r *genMultimapTypeRef) DictName() string {
	return ""
}

func (r *genMultimapTypeRef) DictGoType() string {
	return ""
}

func (r *genMultimapTypeRef) Exported() string {
	return r.TypeName()
}

func (r *genMultimapTypeRef) Storage() string {
	return r.TypeName()
}

func (r *genMultimapTypeRef) ToExported(arg string) string {
	return arg
}

func (r *genMultimapTypeRef) ToStorage(arg string) string {
	return arg
}

func (r *genMultimapTypeRef) TypeName() string {
	str := r.Name
	str = strings.ToUpper(str[:1]) + str[1:]
	return str
}

func (r *genMultimapTypeRef) EncoderType() string {
	return r.TypeName()
}

func (r *genMultimapTypeRef) EqualFunc() string {
	return r.TypeName() + "Equal"
}

func (r *genMultimapTypeRef) CompareFunc() string {
	return "Cmp" + r.TypeName()
}

func (r *genMultimapTypeRef) MustClone() bool {
	return true
}

func (r *genMultimapTypeRef) SchemaStr() string {
	return r.Name
}

func (r *genMultimapTypeRef) Flags() TypeFlags {
	return TypeFlags{
		PassByPtr:  true,
		StoreByPtr: false,
		TakePtr:    true,
	}
}

type genEnumDef struct {
	Name   string
	Fields []*genEnumFieldDef
}

type genEnumFieldDef struct {
	Name  string
	Value uint64
}
