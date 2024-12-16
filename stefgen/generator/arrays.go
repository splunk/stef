package generator

import (
	"fmt"
	"strings"
)

func (g *Generator) oArrays() error {
	var arrays []*genStructFieldDef
	for _, str := range g.compiledSchema.Structs {
		arrays = append(arrays, g.getArrays(str)...)
	}

	for _, mm := range arrays {
		if err := g.oArray(mm); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) getArrays(struc *genStructDef) (ret []*genStructFieldDef) {
	for _, field := range struc.Fields {
		if _, ok := field.Type.(*genArrayTypeRef); ok {
			ret = append(ret, field)
		}
	}
	return ret
}

func (g *Generator) oArray(array *genStructFieldDef) error {
	arrtype := array.Type.(*genArrayTypeRef)

	passByPointer := ""
	if _, ok := arrtype.ElemType.(*genStructTypeRef); ok {
		passByPointer = "&"
	}

	if arrtype.ElemType == nil {
		return fmt.Errorf("array element type unknown")
	}

	_, isStructType := arrtype.ElemType.(*genStructTypeRef)
	data := map[string]any{
		"PackageName":  g.compiledSchema.PackageName,
		"ArrayName":    arrtype.TypeName(),
		"ElemType":     arrtype.ElemType,
		"PassByPtr":    passByPointer,
		"IsStructType": isStructType,
		"Recursive":    array.Recursive,
	}

	return g.oTemplate("array.go.tmpl", strings.ToLower(array.Type.TypeName())+".go", data)
}
