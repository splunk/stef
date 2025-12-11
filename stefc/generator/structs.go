package generator

import (
	"bytes"
	"fmt"
	"go/token"
	"strings"

	"github.com/splunk/stef/go/pkg/schema"
)

func (g *Generator) oStructs() error {
	for _, struc := range g.compiledSchema.Structs {
		err := g.oStruct(struc)
		if err != nil {
			return err
		}
	}

	return g.lastErr
}

func (g *Generator) fieldModifiedConst(str *genStructDef, field *genStructFieldDef) string {
	return fmt.Sprintf("%s%sModified", str.Name, field.Name)
}

func (g *Generator) oStruct(str *genStructDef) error {

	fields := []any{}

	modifier := " = uint64(1 << iota)"
	optionalsMask := uint64(0)
	optionalFieldIndex := 0
	for i, field := range str.Fields {
		passByPointer := false
		if _, ok := field.Type.(*genStructTypeRef); ok {
			passByPointer = true
		}
		if _, ok := field.Type.(*genMultimapTypeRef); ok {
			passByPointer = true
		}
		if _, ok := field.Type.(*genArrayTypeRef); ok {
			passByPointer = true
		}

		_, isStructType := field.Type.(*genStructTypeRef)
		_, isPrimitiveType := field.Type.(*genPrimitiveTypeRef)

		unexportedName := strings.ToLower(field.Name[:1]) + field.Name[1:]
		if token.IsKeyword(unexportedName) {
			unexportedName = unexportedName + "_"
		}

		fieldData := map[string]any{
			"name":          unexportedName,
			"Name":          field.Name,
			"Type":          field.Type,
			"Optional":      field.Optional,
			"FieldIndex":    i,
			"OptionalIndex": optionalFieldIndex,
			"ConstModifier": modifier,
			"PassByPtr":     passByPointer,
			"IsPrimitive":   isPrimitiveType,
			"IsStructType":  isStructType,
		}

		if field.Optional {
			fieldData["OptionalIndex"] = optionalFieldIndex
			optionalFieldIndex++
			optionalsMask |= 1 << uint64(i)
		}

		fields = append(fields, fieldData)

		modifier = ""
	}

	data := map[string]any{
		"StructName":         str.Name,
		"Fields":             fields,
		"DictName":           str.Dict,
		"Type":               str,
		"IsMainStruct":       str.IsRoot,
		"OptionalFieldCount": optionalFieldIndex,
		"OptionalsMask":      optionalsMask,
	}

	if str.IsRoot {
		// Prepare wire schema bytes.
		wireSchema := schema.NewWireSchema(g.schema, str.Name)

		var wireBin bytes.Buffer
		if err := wireSchema.Serialize(&wireBin); err != nil {
			return err
		}

		// Encode as hex string.
		s := ""
		for i, b := range wireBin.Bytes() {
			if i > 0 {
				s += ", "
			}
			s += fmt.Sprintf("0x%02X", b)
		}

		data["Schema"] = s
	}

	templateName := "struct"
	if str.OneOf {
		templateName = "oneof"
	}

	fileName := g.stefSymbol2FileName(str.Name)
	if err := g.oTemplates(templateName, fileName, data); err != nil {
		return err
	}

	return g.lastErr
}

// StructTemplateModel is the data model passed to the struct template to facilitate autocompletion.
type StructTemplateModel struct {
	PackageName string
	StructName  string
	Fields      []StructFieldTemplateModel
}

// StructFieldTemplateModel is the data model passed to the struct template to facilitate autocompletion.
type StructFieldTemplateModel struct {
	name          string
	Name          string
	Type          genFieldTypeRef
	Optional      bool
	FieldIndex    int
	OptionalIndex int
	ConstModifier string
	PassByPtr     bool
	IsPrimitive   bool
	IsStructType  bool
}
