package generator

import (
	"fmt"
	"go/token"
	"strings"
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
	optionalFieldIndex := 0
	for _, field := range str.Fields {
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
			"OptionalIndex": optionalFieldIndex,
			"ConstModifier": modifier,
			"PassByPtr":     passByPointer,
			"IsPrimitive":   isPrimitiveType,
			"IsStructType":  isStructType,
		}

		if field.Optional {
			fieldData["OptionalIndex"] = optionalFieldIndex
			optionalFieldIndex++
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
