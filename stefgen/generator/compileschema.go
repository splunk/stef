package generator

import (
	"fmt"
	"strings"

	"github.com/splunk/stef/go/pkg/schema"
)

func (g *Generator) compileSchema(src *schema.Schema) (*genSchema, error) {
	dst := &genSchema{
		PackageName: src.PackageName,
		Structs:     map[string]*genStructDef{},
		Multimaps:   map[string]*genMapDef{},
		Enums:       map[string]*genEnumDef{},
	}

	switch g.Lang {
	case LangGo:
		// For Go, we use the last component of the package name as the package name.
		dst.PackageNameStr = src.PackageName[len(src.PackageName)-1]
	case LangJava:
		// For Java, we use the full package name.
		dst.PackageNameStr = strings.Join(src.PackageName, ".")
	}

	for name, struc := range src.Structs {
		dst.Structs[name] = structWireToGen(struc, g.Lang)
	}

	for name, multimap := range src.Multimaps {
		dst.Multimaps[name] = multimapWireToGen(multimap, g.Lang)
	}

	for name, enum := range src.Enums {
		dst.Enums[name] = enumSchemaToGen(enum)
	}

	if err := dst.resolveRefs(); err != nil {
		return nil, err
	}

	for _, struc := range dst.Structs {
		if struc.IsRoot {
			stack := recurseStack{asMap: map[string]bool{}}
			computeRecursiveStruct(struc, &stack)
		}
	}

	return dst, nil
}

type recurseStack struct {
	fields  []recursable
	asStack []string
	asMap   map[string]bool
}

func markRecursive(typeName string, stack *recurseStack) {
	startIdx := findLast(stack.asStack, typeName)
	if startIdx == -1 {
		panic("invalid state")
	}
	for i := startIdx; i < len(stack.fields); i++ {
		stack.fields[i].SetRecursive()
	}
}

func computeRecursiveStruct(struc *genStructDef, stack *recurseStack) {
	stack.asStack = append(stack.asStack, struc.Name)
	stack.asMap[struc.Name] = true

	for _, field := range struc.Fields {
		stack.fields = append(stack.fields, field)
		computeRecursiveType(field.Type, stack)
		stack.fields = stack.fields[:len(stack.fields)-1]
	}

	stack.asStack = stack.asStack[:len(stack.asStack)-1]
	delete(stack.asMap, struc.Name)
}

func computeRecursiveMultimap(multimap *genMapDef, stack *recurseStack) {
	stack.asStack = append(stack.asStack, multimap.Name)
	stack.asMap[multimap.Name] = true

	stack.fields = append(stack.fields, &multimap.Key)
	computeRecursiveType(multimap.Key.Type, stack)
	stack.fields = stack.fields[:len(stack.fields)-1]

	stack.fields = append(stack.fields, &multimap.Value)
	computeRecursiveType(multimap.Value.Type, stack)
	stack.fields = stack.fields[:len(stack.fields)-1]

	stack.asStack = stack.asStack[:len(stack.asStack)-1]
	delete(stack.asMap, multimap.Name)
}

func computeRecursiveType(typ genFieldTypeRef, stack *recurseStack) {
	switch t := typ.(type) {
	case *genPrimitiveTypeRef:
		return
	case *genStructTypeRef:
		if stack.asMap[t.Name] {
			markRecursive(t.Name, stack)
		} else {
			computeRecursiveStruct(t.Def, stack)
		}
	case *genMultimapTypeRef:
		if stack.asMap[t.Name] {
			markRecursive(t.Name, stack)
		} else {
			computeRecursiveMultimap(t.Def, stack)
		}
	case *genArrayTypeRef:
		computeRecursiveType(t.ElemType, stack)
	default:
		panic("unknown type")
	}
}

func findLast(stack []string, name string) int {
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i] == name {
			return i
		}
	}
	return -1
}

func (s *genSchema) resolveRefs() error {
	for _, struc := range s.Structs {
		for _, field := range struc.Fields {
			if ref, ok := field.Type.(*genStructTypeRef); ok {
				ref.Def = s.Structs[ref.Name]
				if ref.Def == nil {
					return fmt.Errorf("struct %s not found", ref.Name)
				}
			}
			if ref, ok := field.Type.(*genArrayTypeRef); ok {
				if ref, ok := ref.ElemType.(*genStructTypeRef); ok {
					ref.Def = s.Structs[ref.Name]
					if ref.Def == nil {
						return fmt.Errorf("struct %s not found", ref.Name)
					}
				}
			}
			if ref, ok := field.Type.(*genMultimapTypeRef); ok {
				ref.Def = s.Multimaps[ref.Name]
				if ref.Def == nil {
					return fmt.Errorf("multimap %s not found", ref.Name)
				}
			}
			ref, ok := field.Type.(*genPrimitiveTypeRef)
			if ok && ref.Enum != "" {
				ref.EnumDef = s.Enums[ref.Enum]
			}
		}
	}
	for _, mp := range s.Multimaps {
		if ref, ok := mp.Key.Type.(*genStructTypeRef); ok {
			ref.Def = s.Structs[ref.Name]
			if ref.Def == nil {
				return fmt.Errorf("struct %s not found", ref.Name)
			}
		}
		if ref, ok := mp.Value.Type.(*genStructTypeRef); ok {
			ref.Def = s.Structs[ref.Name]
			if ref.Def == nil {
				return fmt.Errorf("struct %s not found", ref.Name)
			}
		}
		if ref, ok := mp.Key.Type.(*genMultimapTypeRef); ok {
			ref.Def = s.Multimaps[ref.Name]
			if ref.Def == nil {
				return fmt.Errorf("multimap %s not found", ref.Name)
			}
		}
		if ref, ok := mp.Value.Type.(*genMultimapTypeRef); ok {
			ref.Def = s.Multimaps[ref.Name]
			if ref.Def == nil {
				return fmt.Errorf("multimap %s not found", ref.Name)
			}
		}
	}
	return nil
}

func multimapWireToGen(src *schema.Multimap, lang Lang) *genMapDef {
	return &genMapDef{
		Name:  src.Name,
		Key:   multimapFieldWireToAst(src.Key, lang),
		Value: multimapFieldWireToAst(src.Value, lang),
	}
}

func multimapFieldWireToAst(src schema.MultimapField, lang Lang) genMapFieldDef {
	return genMapFieldDef{
		Type: typeWireToGen(src.Type, lang),
		//Recursive: src.Recursive,
	}
}

func structWireToGen(src *schema.Struct, lang Lang) *genStructDef {
	dst := &genStructDef{
		Name:   src.Name,
		OneOf:  src.OneOf,
		IsRoot: src.IsRoot,
		Dict:   src.DictName,
	}

	for i := range src.Fields {
		dst.Fields = append(dst.Fields, structFieldWireToAst(&src.Fields[i], lang))
	}
	return dst
}

func structFieldWireToAst(src *schema.StructField, lang Lang) *genStructFieldDef {
	dst := &genStructFieldDef{
		Name:     src.Name,
		Optional: src.Optional,
	}

	dst.Type = typeWireToGen(src.FieldType, lang)

	return dst
}

func typeWireToGen(src schema.FieldType, lang Lang) genFieldTypeRef {
	if src.Primitive != nil {
		return &genPrimitiveTypeRef{
			Lang: lang,
			Type: *src.Primitive,
			Dict: src.DictName,
			Enum: src.Enum,
		}
	}

	if src.Array != nil {
		return &genArrayTypeRef{
			Lang:     lang,
			ElemType: typeWireToGen(*src.Array, lang),
		}
	}

	if src.Struct != "" {
		return &genStructTypeRef{
			Lang: lang,
			Name: src.Struct,
		}
	}

	if src.MultiMap != "" {
		return &genMultimapTypeRef{
			Lang: lang,
			Name: src.MultiMap,
		}
	}

	panic("unknown field type")
}

func enumSchemaToGen(src *schema.Enum) *genEnumDef {
	dst := &genEnumDef{
		Name: src.Name,
	}
	for i := range src.Fields {
		dst.Fields = append(dst.Fields, enumFieldSchemaToGen(&src.Fields[i]))
	}
	return dst
}

func enumFieldSchemaToGen(src *schema.EnumField) *genEnumFieldDef {
	dst := &genEnumFieldDef{
		Name:  src.Name,
		Value: src.Value,
	}
	return dst
}
