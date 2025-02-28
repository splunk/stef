package idl

import (
	"fmt"

	"github.com/splunk/stef/go/pkg/schema"
)

// Parser parses a STEF IDL input into Schema.
//
// This is a recursive descent parser with separate lexer for tokenization.
type Parser struct {
	lexer    *Lexer
	schema   *schema.Schema
	fileName string
}

// Error represents a parsing error.
type Error struct {
	Msg      string
	Filename string
	Pos      Pos
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.Filename, e.Pos.Line, e.Pos.Col, e.Msg)
}

var _ error = (*Error)(nil)

// NewParser creates a new parser with specified lexer as the input.
// fileName is used for composing error messages (if any).
func NewParser(lexer *Lexer, fileName string) *Parser {
	p := &Parser{fileName: fileName}
	p.lexer = lexer
	p.schema = &schema.Schema{}
	return p
}

// Schema returns the parsed Schema, assuming Parse() returned nil.
func (p *Parser) Schema() *schema.Schema {
	return p.schema
}

// Parse an IDL input into Schema.
// Will return an error if the input syntax is invalid.
func (p *Parser) Parse() error {
	p.schema = &schema.Schema{
		Structs:   map[string]*schema.Struct{},
		Multimaps: map[string]*schema.Multimap{},
		Enums:     map[string]*schema.Enum{},
	}

	if err := p.parsePackage(); err != nil {
		return err
	}

	for {
		var err error
		switch p.lexer.Token() {
		case tStruct:
			_, err = p.parseStruct()
		case tOneof:
			err = p.parseOneof()
		case tMultimap:
			err = p.parseMultimap()
		case tEnum:
			err = p.parseEnum()
		default:
			return p.error("expected struct, oneof or multimap")
		}
		if err != nil {
			return err
		}
		if p.lexer.Token() == tEOF {
			break
		}
	}
	return p.resolveFieldTypes()
}

func (p *Parser) isTopLevelNameUsed(name string) bool {
	return p.schema.Structs[name] != nil || p.schema.Multimaps[name] != nil || p.schema.Enums[name] != nil
}

func (p *Parser) parseStruct() (*schema.Struct, error) {
	p.lexer.Next() // skip "struct"

	if p.lexer.Token() != tIdent {
		return nil, p.error("struct name expected")
	}
	structName := p.lexer.Ident()

	if p.isTopLevelNameUsed(structName) {
		return nil, p.error("duplicate top-level identifier: " + structName)
	}

	p.lexer.Next()

	str := &schema.Struct{
		Name: structName,
	}
	p.schema.Structs[str.Name] = str

	if err := p.parseStructModifiers(str); err != nil {
		return nil, err
	}

	if err := p.eat(tLBrace); err != nil {
		return nil, err
	}

	if err := p.parseStructFields(str); err != nil {
		return nil, err
	}

	if err := p.eat(tRBrace); err != nil {
		return nil, err
	}

	return str, nil
}

func (p *Parser) parseOneof() error {
	// "oneof" syntax is identical to struct, except we need to set "OneOf" flag.
	str, err := p.parseStruct()
	if err != nil {
		return err
	}
	str.OneOf = true
	return nil
}

func (p *Parser) parseMultimap() error {
	p.lexer.Next() // skip "multimap"

	if p.lexer.Token() != tIdent {
		return p.error("multimap name expected")
	}
	multimapName := p.lexer.Ident()

	if p.isTopLevelNameUsed(multimapName) {
		return p.error("duplicate top-level identifier: " + multimapName)
	}

	p.lexer.Next()

	mm := &schema.Multimap{
		Name: multimapName,
	}
	p.schema.Multimaps[mm.Name] = mm

	if err := p.eat(tLBrace); err != nil {
		return err
	}

	// Parse the key.
	if err := p.eat(tKey); err != nil {
		return err
	}
	if err := p.parseMultimapField(&mm.Key); err != nil {
		return err
	}

	// Parse the value.
	if err := p.eat(tValue); err != nil {
		return err
	}
	if err := p.parseMultimapField(&mm.Value); err != nil {
		return err
	}

	if err := p.eat(tRBrace); err != nil {
		return err
	}

	return nil
}

func (p *Parser) error(msg string) error {
	return &Error{
		Msg:      msg,
		Filename: p.fileName,
		Pos:      p.lexer.TokenStartPos(),
	}
}

func (p *Parser) parseStructModifiers(str *schema.Struct) error {
	for {
		err, ok := p.parseStructModifier(str)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
	}
	return nil
}

func (p *Parser) parseStructModifier(str *schema.Struct) (error, bool) {
	switch p.lexer.Token() {
	case tDict:
		dictName, err := p.parseDictModifier()
		if err != nil {
			return err, false
		}
		str.DictName = dictName
	case tRoot:
		str.IsRoot = true
		p.lexer.Next()
	default:
		return nil, false
	}
	return nil, false
}

func (p *Parser) parseDictModifier() (string, error) {
	p.lexer.Next() // skip "dict"

	if err := p.eat(tLParen); err != nil {
		return "", err
	}

	if p.lexer.Token() != tIdent {
		return "", p.error("dict name expected")
	}
	dictName := p.lexer.Ident()
	p.lexer.Next()

	if err := p.eat(tRParen); err != nil {
		return "", err
	}
	return dictName, nil
}

// eat checks that the current token is the expected one and skips it.
func (p *Parser) eat(token Token) error {
	if p.lexer.Token() != token {
		return p.error(fmt.Sprintf("expected %s but got %s", token, p.lexer.Token()))
	}
	p.lexer.Next()
	return nil
}

func (p *Parser) parseStructFields(str *schema.Struct) error {
	for {
		err, ok := p.parseStructField(str)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
	}
	return nil
}

func (p *Parser) parseStructField(str *schema.Struct) (error, bool) {
	if p.lexer.Token() != tIdent {
		return nil, false
	}

	str.Fields = append(str.Fields, schema.StructField{Name: p.lexer.Ident()})
	field := &str.Fields[len(str.Fields)-1]

	p.lexer.Next()

	if err := p.parseFieldType(&field.FieldType); err != nil {
		return err, false
	}
	if err := p.parseStructFieldModifiers(field); err != nil {
		return err, false
	}

	return nil, true
}

func (p *Parser) parseFieldType(field *schema.FieldType) error {
	isArray := false
	if p.lexer.Token() == tLBracket {
		isArray = true
		p.lexer.Next()
		// We expect a matching right bracket.
		if err := p.eat(tRBracket); err != nil {
			return err
		}
	}

	ft := schema.FieldType{}
	switch p.lexer.Token() {
	case tIdent:
		// Temporarily store in "Struct", but this may also be a oneof or multimap.
		// We will resolve to the correct type it later, after all input is read,
		// since it may be a forward reference.
		ft.Struct = p.lexer.Ident()

	case tBool:
		v := schema.PrimitiveTypeBool
		ft.Primitive = &v

	case tInt64:
		v := schema.PrimitiveTypeInt64
		ft.Primitive = &v

	case tUint64:
		v := schema.PrimitiveTypeUint64
		ft.Primitive = &v

	case tFloat64:
		v := schema.PrimitiveTypeFloat64
		ft.Primitive = &v

	case tString:
		v := schema.PrimitiveTypeString
		ft.Primitive = &v

	case tBytes:
		v := schema.PrimitiveTypeBytes
		ft.Primitive = &v

	default:
		if isArray {
			return p.error("type specifier expected after []")
		}
		return nil
	}
	p.lexer.Next()

	if isArray {
		field.Array = &ft
	} else {
		*field = ft
	}

	return nil
}

func (p *Parser) parseStructFieldModifiers(field *schema.StructField) error {
	for {
		err, ok := p.parseStructFieldModifier(field)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
	}
	return nil
}

func (p *Parser) parseStructFieldModifier(field *schema.StructField) (error, bool) {
	switch p.lexer.Token() {
	case tDict:
		dictName, err := p.parseDictModifier()
		if err != nil {
			return err, false
		}
		field.DictName = dictName
		return nil, true
	case tOptional:
		field.Optional = true
		p.lexer.Next()
		return nil, true
	default:
		return nil, false
	}
}

func (p *Parser) parseMultimapField(field *schema.MultimapField) error {
	if err := p.parseFieldType(&field.Type); err != nil {
		return err
	}

	if p.lexer.Token() == tDict {
		dictName, err := p.parseDictModifier()
		if err != nil {
			return err
		}
		field.Type.DictName = dictName
	}

	return nil
}

func (p *Parser) resolveFieldTypes() error {
	for _, v := range p.schema.Structs {
		for i := range v.Fields {
			field := &v.Fields[i]
			if err := p.resolveFieldType(&field.FieldType); err != nil {
				return err
			}
		}
	}
	for _, v := range p.schema.Multimaps {
		if err := p.resolveFieldType(&v.Key.Type); err != nil {
			return err
		}
		if err := p.resolveFieldType(&v.Value.Type); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) resolveFieldType(fieldType *schema.FieldType) error {
	typeName := fieldType.Struct
	if typeName != "" {
		matches := 0
		_, isStruct := p.schema.Structs[typeName]
		if isStruct {
			matches++
		}

		_, isMultimap := p.schema.Multimaps[typeName]
		if isMultimap {
			fieldType.MultiMap = typeName
			fieldType.Struct = ""
			matches++
		}

		_, isEnum := p.schema.Enums[typeName]
		if isEnum {
			// All enums are uint64.
			t := schema.PrimitiveTypeUint64
			fieldType.Primitive = &t
			fieldType.Enum = typeName
			fieldType.Struct = ""
			matches++
		}

		if matches == 0 {
			return p.error("unknown type: " + typeName)
		}
		if matches > 1 {
			return p.error("ambiguous type: " + typeName)
		}
	}
	return nil
}

func (p *Parser) parsePackage() error {
	if p.lexer.Token() == tPackage {
		p.lexer.Next() // skip "package"

		if p.lexer.Token() != tIdent {
			return p.error("package name expected")
		}
		p.schema.PackageName = p.lexer.Ident()
		p.lexer.Next()
	}
	return nil
}

func (p *Parser) parseEnum() error {
	p.lexer.Next() // skip "enum"

	if p.lexer.Token() != tIdent {
		return p.error("enum name expected")
	}
	enumName := p.lexer.Ident()

	if p.isTopLevelNameUsed(enumName) {
		return p.error("duplicate top-level identifier: " + enumName)
	}

	p.lexer.Next()

	enum := &schema.Enum{
		Name: enumName,
	}
	p.schema.Enums[enum.Name] = enum

	if err := p.eat(tLBrace); err != nil {
		return err
	}

	if err := p.parseEnumFields(enum); err != nil {
		return err
	}

	if err := p.eat(tRBrace); err != nil {
		return err
	}

	return nil
}

func (p *Parser) parseEnumFields(enum *schema.Enum) error {
	for {
		err, ok := p.parseEnumField(enum)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
	}
	return nil
}

func (p *Parser) parseEnumField(enum *schema.Enum) (error, bool) {
	if p.lexer.Token() != tIdent {
		return nil, false
	}

	enum.Fields = append(enum.Fields, schema.EnumField{Name: p.lexer.Ident()})
	field := &enum.Fields[len(enum.Fields)-1]

	p.lexer.Next() // skip field name

	if err := p.eat(tAssign); err != nil {
		return err, false
	}

	if p.lexer.Token() != tIntNumber {
		errMsg := "enum field value expected"
		if p.lexer.Token() == tError {
			errMsg += ": " + p.lexer.ErrMsg()
		}
		return p.error(errMsg), false
	}

	field.Value = p.lexer.Uint64Number()
	p.lexer.Next()

	return nil, true
}
