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
	MainStruct  string
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
		if struc.Name == s.MainStruct {
			str += " main"
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
	// PassByPtr indicates that the value of this type is passed by pointer to GoQualifiedType()
	// when it is a function parameter or when it is returned by a function.
	PassByPtr bool

	// StoreByPtr indicates that struct fields of the value of this type is stored as a
	// pointer to GoQualifiedType(). If this is false that the fields are simply of GoQualifiedType().
	StoreByPtr bool

	// TakePtr is true a pointer must be taken of the field to pass it as a parameter.
	TakePtr bool
}

type genFieldTypeRef interface {
	TypeName() string
	IsPrimitive() bool

	// GoQualifiedType is the fully qualified Go type for this field type.
	GoQualifiedType() string
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
	Type schema.PrimitiveFieldType
	Dict string
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

func (r *genPrimitiveTypeRef) GoQualifiedType() string {
	if r.Type == schema.PrimitiveTypeBytes {
		return "pkg.Bytes"
	}
	return r.TypeName()
}

func (r *genPrimitiveTypeRef) TypeName() string {
	var s string

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

	return s
}

func (r *genPrimitiveTypeRef) EncoderType() string {
	switch r.Type {
	case schema.PrimitiveTypeUint64:
		return "encoders.Uint64"
	case schema.PrimitiveTypeInt64:
		return "encoders.Int64"
	case schema.PrimitiveTypeFloat64:
		return "encoders.Float64"
	case schema.PrimitiveTypeBool:
		return "encoders.Bool"
	case schema.PrimitiveTypeString:
		return "encoders.String"
	case schema.PrimitiveTypeBytes:
		return "encoders.Bytes"
	default:
		panic(fmt.Sprintf("unknown type %v", r.Type))
	}
}

func (r *genPrimitiveTypeRef) DictGoType() string {
	switch r.Type {
	case schema.PrimitiveTypeString:
		return "encoders.String"
	case schema.PrimitiveTypeBytes:
		return "encoders.Bytes"
	default:
		panic(fmt.Sprintf("type %v does not support dictionaries", r.Type))
	}
}

func (r *genPrimitiveTypeRef) EqualFunc() string {
	switch r.Type {
	case schema.PrimitiveTypeUint64:
		return "pkg.Uint64Equal"
	case schema.PrimitiveTypeInt64:
		return "pkg.Int64Equal"
	case schema.PrimitiveTypeFloat64:
		return "pkg.Float64Equal"
	case schema.PrimitiveTypeBool:
		return "pkg.BoolEqual"
	case schema.PrimitiveTypeString:
		return "pkg.StringEqual"
	case schema.PrimitiveTypeBytes:
		return "pkg.BytesEqual"
	default:
		panic(fmt.Sprintf("unknown type %v", r.Type))
	}
}

func (r *genPrimitiveTypeRef) RandomFunc() string {
	switch r.Type {
	case schema.PrimitiveTypeUint64:
		return "pkg.Uint64Random"
	case schema.PrimitiveTypeInt64:
		return "pkg.Int64Random"
	case schema.PrimitiveTypeFloat64:
		return "pkg.Float64Random"
	case schema.PrimitiveTypeBool:
		return "pkg.BoolRandom"
	case schema.PrimitiveTypeString:
		return "pkg.StringRandom"
	case schema.PrimitiveTypeBytes:
		return "pkg.BytesRandom"
	default:
		panic(fmt.Sprintf("unknown type %v", r.Type))
	}
}

func (r *genPrimitiveTypeRef) CompareFunc() string {
	switch r.Type {
	case schema.PrimitiveTypeUint64:
		return "pkg.Uint64Compare"
	case schema.PrimitiveTypeInt64:
		return "pkg.Int64Compare"
	case schema.PrimitiveTypeFloat64:
		return "pkg.Float64Compare"
	case schema.PrimitiveTypeBool:
		return "pkg.BoolCompare"
	case schema.PrimitiveTypeString:
		return "strings.Compare"
	case schema.PrimitiveTypeBytes:
		return "pkg.BytesCompare"
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

func (r *genStructTypeRef) GoQualifiedType() string {
	return r.TypeName()
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
	return "Cmp" + r.Name
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

func (r *genArrayTypeRef) GoQualifiedType() string {
	return r.TypeName()
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
	return "Cmp" + r.TypeName()
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

func (r *genMultimapTypeRef) GoQualifiedType() string {
	return r.TypeName()
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
