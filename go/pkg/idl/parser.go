package idl

import (
	"errors"
	"fmt"

	"github.com/splunk/stef/go/pkg/schema"
)

type Parser struct {
	lexer  *Lexer
	schema *schema.Schema
}

func NewParser(l *Lexer) *Parser {
	p := &Parser{}
	p.lexer = l
	p.schema = &schema.Schema{}
	l.Start()
	return p
}

func (p *Parser) Schema() *schema.Schema {
	return p.schema
}

func (p *Parser) Parse() error {
	p.schema = &schema.Schema{
		Structs:   map[string]*schema.Struct{},
		Multimaps: map[string]*schema.Multimap{},
	}

	for {
		var err error
		switch p.lexer.Token() {
		case tStruct:
			_, err = p.struc()
		case tOneof:
			err = p.oneof()
		case tMultimap:
			err = p.multimap()
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
	return nil
}

func (p *Parser) struc() (*schema.Struct, error) {
	p.lexer.Next()
	if p.lexer.Token() != tIdent {
		return nil, p.error("struct name expected")
	}
	structName := p.lexer.Ident()
	p.lexer.Next()

	str := &schema.Struct{
		Name: structName,
	}
	p.schema.Structs[str.Name] = str

	if err := p.structModifiers(str); err != nil {
		return nil, err
	}
	if err := p.eat(tLBrace); err != nil {
		return nil, err
	}

	if err := p.structFields(str); err != nil {
		return nil, err
	}

	if err := p.eat(tRBrace); err != nil {
		return nil, err
	}

	return str, nil
}

func (p *Parser) oneof() error {
	str, err := p.struc()
	if err != nil {
		return err
	}
	str.OneOf = true
	return nil
}

func (p *Parser) multimap() error {
	p.lexer.Next()
	if p.lexer.Token() != tIdent {
		return p.error("multimap name expected")
	}
	multimapName := p.lexer.Ident()
	p.lexer.Next()

	mm := &schema.Multimap{
		Name: multimapName,
	}
	p.schema.Multimaps[mm.Name] = mm

	if err := p.eat(tLBrace); err != nil {
		return err
	}

	if err := p.eat(tKey); err != nil {
		return err
	}

	if err := p.multimapField(&mm.Key); err != nil {
		return err
	}

	if err := p.eat(tValue); err != nil {
		return err
	}
	if err := p.multimapField(&mm.Value); err != nil {
		return err
	}

	if err := p.eat(tRBrace); err != nil {
		return err
	}

	return nil
}

func (p *Parser) error(msg string) error {
	return errors.New(msg)
}

func (p *Parser) structModifiers(str *schema.Struct) error {
	for {
		err, ok := p.structModifier(str)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
	}
	return nil
}

func (p *Parser) structModifier(str *schema.Struct) (error, bool) {
	switch p.lexer.Token() {
	case tDict:
		if dictName, err := p.dictModifier(); err != nil {
			str.DictName = dictName
		}
	case tRoot:
		str.IsRoot = true
		p.lexer.Next()
	}
	return nil, false
}

func (p *Parser) dictModifier() (string, error) {
	p.lexer.Next()
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

func (p *Parser) eat(token Token) error {
	if p.lexer.Token() != token {
		return fmt.Errorf("expected %s but got %s", token, p.lexer.Token())
	}
	p.lexer.Next()
	return nil
}

func (p *Parser) structFields(str *schema.Struct) error {
	for {
		err, ok := p.structField(str)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
	}
	return nil
}

func (p *Parser) structField(str *schema.Struct) (error, bool) {
	if p.lexer.Token() != tIdent {
		return p.error("field name expected"), false
	}
	p.lexer.Next()

	str.Fields = append(str.Fields, schema.StructField{Name: p.lexer.Ident()})
	field := &str.Fields[len(str.Fields)-1]

	if err := p.fieldType(&field.FieldType); err != nil {
		return err, false
	}
	if err := p.structFieldModifiers(field); err != nil {
		return err, false
	}

	return nil, true
}

func (p *Parser) fieldType(field *schema.FieldType) error {
	isArray := false
	if p.lexer.Token() == tLBracket {
		isArray = true
		p.lexer.Next()
		if err := p.eat(tRBracket); err != nil {
			return err
		}
	}

	ft := schema.FieldType{}
	switch p.lexer.Token() {
	case tIdent:
		// Temporarily store in "Struct", but this may also be a oneof or multimap.
		// We will resolve it later.
		ft.Struct = p.lexer.Ident()

	case tBool:
		v := schema.PrimitiveTypeBool
		ft.Primitive = &v

	case tInt64:
		v := schema.PrimitiveTypeInt64
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

func (p *Parser) structFieldModifiers(field *schema.StructField) error {
	for {
		err, ok := p.structFieldModifier(field)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
	}
	return nil
}

func (p *Parser) structFieldModifier(field *schema.StructField) (error, bool) {
	switch p.lexer.Token() {
	case tDict:
		if dictName, err := p.dictModifier(); err != nil {
			field.DictName = dictName
		}
		return nil, true
	case tOptional:
		field.Optional = true
		return nil, true
	}
	return nil, false
}

func (p *Parser) multimapField(field *schema.MultimapField) error {
	if err := p.fieldType(&field.Type); err != nil {
		return err
	}

	if p.lexer.Token() == tDict {
		dictName, err := p.dictModifier()
		if err != nil {
			return err
		}
		field.Type.DictName = dictName
	}

	return nil
}
