package generator

import (
	"fmt"
)

func (g *Generator) oArrays() error {
	var arrFields []*genFieldDef

	// Collect all array fields from all structs
	for _, str := range g.compiledSchema.Structs {
		arrFields = append(arrFields, g.getArrays(str)...)
	}

	// Collect all array fields from all multimaps (key and value)
	for _, mm := range g.compiledSchema.Multimaps {
		if _, ok := mm.Key.Type.(*genArrayTypeRef); ok {
			arrFields = append(arrFields, &mm.Key.genFieldDef)
		}
		if _, ok := mm.Value.Type.(*genArrayTypeRef); ok {
			arrFields = append(arrFields, &mm.Value.genFieldDef)
		}
	}

	for _, arrField := range arrFields {
		if err := g.oArray(arrField); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) getArrays(struc *genStructDef) (ret []*genFieldDef) {
	for _, field := range struc.Fields {
		if _, ok := field.Type.(*genArrayTypeRef); ok {
			ret = append(ret, &field.genFieldDef)
		}
	}
	return ret
}

func (g *Generator) oArray(fieldDef *genFieldDef) error {
	arrtype := fieldDef.Type.(*genArrayTypeRef)

	passByPointer := ""
	if _, ok := arrtype.ElemType.(*genStructTypeRef); ok {
		passByPointer = "&"
	}

	if arrtype.ElemType == nil {
		return fmt.Errorf("array element type unknown")
	}

	_, isStructType := arrtype.ElemType.(*genStructTypeRef)
	data := map[string]any{
		"PackageName":  g.compiledSchema.PackageNameStr,
		"ArrayName":    arrtype.IDLMangledName(),
		"ElemType":     arrtype.ElemType,
		"PassByPtr":    passByPointer,
		"IsStructType": isStructType,
		"Recursive":    fieldDef.Recursive,
	}

	return g.oTemplates("array", g.stefSymbol2FileName(arrtype.IDLMangledName()), data)
}
