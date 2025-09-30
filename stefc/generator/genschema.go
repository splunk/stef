package generator

import (
	"fmt"
	"strings"

	"github.com/splunk/stef/go/pkg/schema"
)

// genSchema is a STEF schema description in form that is useful for generation.
type genSchema struct {
	PackageName    []string
	PackageNameStr string
	Structs        map[string]*genStructDef
	Multimaps      map[string]*genMapDef
	Enums          map[string]*genEnumDef
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

type genStructFieldDef struct {
	Name      string
	Type      genFieldTypeRef
	Optional  bool
	Recursive bool
}

type TypeFlags struct {
	// PassByPtr indicates that the value of this type is passed by pointer to Exported()
	// when it is a function parameter or when it is returned by a function.
	PassByPtr bool

	// StoreByPtr indicates that struct fields of the value of this type is stored as a
	// pointer to Exported(). If this is false that the fields are simply of Exported().
	StoreByPtr bool

	// TakePtr is true when a pointer must be taken of the field to pass it as a parameter when encoding or comparing.
	TakePtr bool

	// DecodeByPtr is true when a pointer must be taken of the field to pass it as a parameter to Decode().
	DecodeByPtr bool

	IsEnum bool
}

type genFieldTypeRef interface {
	TypeName() string
	IsPrimitive() bool

	// IDLMangledName returns the name of the type derived from its STEF IDL
	// declaration mangled into an identifier that uniquely identifies the type
	// and is used as the identifier of that type throughout the generated code.
	IDLMangledName() string

	// Exported is the fully qualified exported (public) type.
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

	// DictTypeNamePrefix is the prefix for fully qualified type name for encoder and
	// decoder dictionaries.
	DictTypeNamePrefix() string

	SchemaStr() string

	Flags() TypeFlags
}

type genPrimitiveTypeRef struct {
	Type schema.PrimitiveFieldType

	// Dict is the name of the dictionary type if this is a dictionary.
	Dict string

	// Indicates if "delta" modifier is applied.
	Delta schema.DeltaModifier

	// Enum is the name of the enum type if this is an enum.
	Enum string
	// EnumDef is the definition of the enum type.
	EnumDef *genEnumDef

	// Lang is the language for which code is being generated.
	// Lang is used to correctly generate language-specific code.
	Lang Lang
}

func (r *genPrimitiveTypeRef) IsPrimitive() bool {
	return true
}

func (r *genPrimitiveTypeRef) Flags() TypeFlags {
	isEnum := false
	if r.Enum != "" {
		isEnum = true
	}
	return TypeFlags{
		PassByPtr:   false,
		StoreByPtr:  false,
		TakePtr:     false,
		DecodeByPtr: true,
		IsEnum:      isEnum,
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
			return "byte[]"
		default:
			panic(fmt.Sprintf("unknown language %v", r.Lang))
		}
	}
	return r.TypeName()
}

// InitVal returns the initial value for this type.
// Return empty string if there is no need to assign an initial value
// since the default initial value is good enough.
func (r *genPrimitiveTypeRef) InitVal() string {
	switch r.Type {
	case schema.PrimitiveTypeInt64:
		return "0"
	case schema.PrimitiveTypeUint64:
		return "0"
	case schema.PrimitiveTypeFloat64:
		return "0.0"
	case schema.PrimitiveTypeBool:
		return "false"
	case schema.PrimitiveTypeString:
		switch r.Lang {
		case LangGo:
			return `""`
		case LangJava:
			return "StringValue.empty"
		default:
			panic("unknown language")
		}
	case schema.PrimitiveTypeBytes:
		switch r.Lang {
		case LangGo:
			return "pkg.EmptyBytes"
		case LangJava:
			return "Types.emptyBytes"
		default:
			panic("unknown language")
		}
	default:
		panic("unknown type")
	}
}

// ToStorage converts the argument to the underlying type
// if the underlying type is different than the exported type.
// If the types are the same, no conversion is performed.
func (r *genPrimitiveTypeRef) ToStorage(arg string) string {
	if r.Enum != "" {
		switch r.Lang {
		case LangGo:
			return r.Storage() + "(" + arg + ")"
		case LangJava:
			return arg + ".getValue()"
		default:
			panic(fmt.Sprintf("unknown language %v", r.Lang))
		}
	}
	return arg
}

// ToExported converts the argument to the exported type
// if the underlying storage type is different than the exported type.
// If the types are the same, no conversion is performed.
func (r *genPrimitiveTypeRef) ToExported(arg string) string {
	if r.Enum != "" {
		switch r.Lang {
		case LangGo:
			return r.Enum + "(" + arg + ")"
		case LangJava:
			return r.Enum + ".fromValue(" + arg + ")"
		default:
			panic(fmt.Sprintf("unknown language %v", r.Lang))
		}
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
	return r.Storage()
}

// TypeName returns the language-specific type of the primitive type.
func (r *genPrimitiveTypeRef) TypeName() string {
	var typeMap map[schema.PrimitiveFieldType]string

	switch r.Lang {
	case LangGo:
		typeMap = map[schema.PrimitiveFieldType]string{
			schema.PrimitiveTypeInt64:   "int64",
			schema.PrimitiveTypeUint64:  "uint64",
			schema.PrimitiveTypeFloat64: "float64",
			schema.PrimitiveTypeBool:    "bool",
			schema.PrimitiveTypeString:  "string",
			schema.PrimitiveTypeBytes:   "Bytes",
		}
	case LangJava:
		typeMap = map[schema.PrimitiveFieldType]string{
			schema.PrimitiveTypeInt64:   "long",
			schema.PrimitiveTypeUint64:  "long",
			schema.PrimitiveTypeFloat64: "double",
			schema.PrimitiveTypeBool:    "boolean",
			schema.PrimitiveTypeString:  "StringValue",
			schema.PrimitiveTypeBytes:   "byte[]",
		}
	default:
		panic(fmt.Sprintf("unknown language %v", r.Lang))
	}

	if s, ok := typeMap[r.Type]; ok {
		return s
	}
	panic(fmt.Errorf("unimplemented field type %v", r.Type))
}

// Names of primitive types as used in conventions for names of functions or
// other symbols for that particular primitive type (e.g. in Uint64Equal function name).
var primitiveTypeMangledNames = map[schema.PrimitiveFieldType]string{
	schema.PrimitiveTypeUint64:  "Uint64",
	schema.PrimitiveTypeInt64:   "Int64",
	schema.PrimitiveTypeFloat64: "Float64",
	schema.PrimitiveTypeBool:    "Bool",
	schema.PrimitiveTypeString:  "String",
	schema.PrimitiveTypeBytes:   "Bytes",
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

	if s, ok := primitiveTypeMangledNames[r.Type]; ok {
		name := prefix + s // e.g. encoders.Uint64
		if r.Dict != "" {
			name += "Dict"
		}
		switch r.Delta {
		case schema.DeltaModifierNone:
		case schema.DeltaModifierDelta:
			name += "Delta"
		case schema.DeltaModifierDeltaDelta:
			name += "DeltaDelta"
		default:
			panic(fmt.Sprintf("unknown delta modifier %v", r.Delta))
		}
		return name
	}

	panic(fmt.Sprintf("unknown type %v", r.Type))
}

func (r *genPrimitiveTypeRef) IDLMangledName() string {
	if s, ok := primitiveTypeMangledNames[r.Type]; ok {
		return s // e.g. Uint64
	}
	panic(fmt.Sprintf("unknown type %v", r.Type))
}

// DictTypeNamePrefix is the prefix for fully qualified type name for encoder and
// decoder dictionaries.
func (r *genPrimitiveTypeRef) DictTypeNamePrefix() string {
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
		return prefix + "StringDict"
	case schema.PrimitiveTypeBytes:
		return prefix + "BytesDict"
	default:
		panic(fmt.Sprintf("type %v does not support dictionaries", r.Type))
	}
}

// pkgPrefix returns the language-specific supporting package name.
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
	prefix := r.pkgPrefix()

	if s, ok := primitiveTypeMangledNames[r.Type]; ok {
		return prefix + s + "Equal" // e.g. Uint64Equal
	}

	panic(fmt.Sprintf("unknown type %v", r.Type))
}

func (r *genPrimitiveTypeRef) RandomFunc() string {
	prefix := r.pkgPrefix()

	if r.Enum != "" {
		// For enums we need to generate an unsigned random number in the range of the enum values.
		switch r.Lang {
		case LangGo:
			return r.Enum + fmt.Sprintf("(pkg.Uint64Random(random) %% %d)", len(r.EnumDef.Fields))
		case LangJava:
			return r.Enum + fmt.Sprintf(
				".fromValue((Types.Uint64Random(random) & 0x7FFFFFFFFFFFFFFFL) %% %d)", len(r.EnumDef.Fields),
			)
		default:
			panic(fmt.Sprintf("unknown language %v", r.Lang))
		}
	}

	if s, ok := primitiveTypeMangledNames[r.Type]; ok {
		return prefix + s + "Random(random)" // e.g. Uint64Random(random)
	}

	panic(fmt.Sprintf("unknown type %v", r.Type))
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
		if r.Lang == LangGo {
			return "strings.Compare"
		}
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

// SchemaStr returns type names in STEF IDL syntax.
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
	Def    *schema.Struct
	Name   string
	Dict   string
	Fields []*genStructFieldDef
	OneOf  bool
	IsRoot bool
}

func (r *genStructDef) Flags() TypeFlags {
	storeByPtr := r.Dict != "" || // dictionary-encoded structs are stored by pointer to dictionary element
		r.Def.Recursive() // recursive structs are stored by pointer to break the recursion

	return TypeFlags{
		PassByPtr:  true,
		StoreByPtr: storeByPtr,
		TakePtr:    !storeByPtr,
		// Recursive structs are already stored by pointer, no need to take pointer of that.
		// See implementations of Decode() methods.
		DecodeByPtr: !r.Def.Recursive(),
	}
}

type genMapFieldDef struct {
	Type      genFieldTypeRef
	Recursive bool
}

type genMapDef struct {
	Name  string
	Key   genMapFieldDef
	Value genMapFieldDef
}

type genStructTypeRef struct {
	Name string
	Def  *genStructDef
	// Lang is the language for which code is being generated.
	// Lang is used to correctly generate language-specific code.
	Lang Lang
}

func (r *genStructTypeRef) IsPrimitive() bool {
	return false
}

func (r *genStructTypeRef) DictName() string {
	return r.Def.Dict
}

func (r *genStructTypeRef) DictTypeNamePrefix() string {
	return r.DictName()
}

func (r *genStructTypeRef) Exported() string {
	return r.Name
}

func (r *genStructTypeRef) Storage() string {
	return r.Name
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

func (r *genStructTypeRef) IDLMangledName() string {
	return r.Name
}

func (r *genStructTypeRef) EncoderType() string {
	return r.Name
}

func (r *genStructTypeRef) EqualFunc() string {
	switch r.Lang {
	case LangGo:
		return r.Name + "Equal"
	case LangJava:
		return r.Name + ".equals"
	default:
		panic(fmt.Sprintf("unknown language %v", r.Lang))
	}
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
	return r.Def.Flags()
}

type genArrayTypeRef struct {
	ElemType genFieldTypeRef
	// Lang is the language for which code is being generated.
	// Lang is used to correctly generate language-specific code.
	Lang Lang
}

func (r *genArrayTypeRef) Flags() TypeFlags {
	return TypeFlags{
		PassByPtr:   true,
		StoreByPtr:  false,
		TakePtr:     true,
		DecodeByPtr: true,
	}
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

func (r *genArrayTypeRef) DictTypeNamePrefix() string {
	return r.ElemType.DictTypeNamePrefix()
}

func (r *genArrayTypeRef) Exported() string {
	return r.IDLMangledName()
}

func (r *genArrayTypeRef) Storage() string {
	return r.IDLMangledName()
}

func (r *genArrayTypeRef) ToExported(arg string) string {
	return r.ToExported(arg)
}

func (r *genArrayTypeRef) ToStorage(arg string) string {
	return arg
}

func (r *genArrayTypeRef) TypeName() string {
	str := r.ElemType.TypeName() + "Array"
	// Make sure the type name is exported.
	str = strings.ToUpper(str[:1]) + str[1:]
	return str
}

func (r *genArrayTypeRef) EncoderType() string {
	return r.IDLMangledName()
}

func (r *genArrayTypeRef) EqualFunc() string {
	return r.TypeName() + "Equal"
}

func (r *genArrayTypeRef) CompareFunc() string {
	switch r.Lang {
	case LangGo:
		return "Cmp" + r.IDLMangledName()
	case LangJava:
		return r.IDLMangledName() + ".compare"
	default:
		panic(fmt.Sprintf("unknown language %v", r.Lang))
	}
}

func (r *genArrayTypeRef) MustClone() bool {
	return true
}

func (r *genArrayTypeRef) IDLMangledName() string {
	return r.ElemType.IDLMangledName() + "Array"
}

type genMultimapTypeRef struct {
	Name string
	Def  *genMapDef

	// Lang is the language for which code is being generated.
	// Lang is used to correctly generate language-specific code.
	Lang Lang
}

func (r *genMultimapTypeRef) IsPrimitive() bool {
	return false
}

func (r *genMultimapTypeRef) DictName() string {
	return ""
}

func (r *genMultimapTypeRef) DictTypeNamePrefix() string {
	panic("dictionaries of multimaps are not supported")
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

func (r *genMultimapTypeRef) IDLMangledName() string {
	return r.TypeName()
}

func (r *genMultimapTypeRef) EncoderType() string {
	return r.TypeName()
}

func (r *genMultimapTypeRef) EqualFunc() string {
	return r.TypeName() + "Equal"
}

func (r *genMultimapTypeRef) CompareFunc() string {
	switch r.Lang {
	case LangGo:
		return "Cmp" + r.TypeName()
	case LangJava:
		return r.Name + ".compare"
	default:
		panic(fmt.Sprintf("unknown language %v", r.Lang))
	}
}

func (r *genMultimapTypeRef) MustClone() bool {
	return true
}

func (r *genMultimapTypeRef) SchemaStr() string {
	return r.Name
}

func (r *genMultimapTypeRef) Flags() TypeFlags {
	return TypeFlags{
		PassByPtr:   true,
		StoreByPtr:  false,
		TakePtr:     true,
		DecodeByPtr: true,
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
