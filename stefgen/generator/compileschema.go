package generator

import (
	"fmt"

	"github.com/splunk/stef/go/pkg/schema"
)

func compileSchema(src *schema.Schema) (*genSchema, error) {
	dst := &genSchema{
		PackageName: src.PackageName,
		Structs:     map[string]*genStructDef{},
		Multimaps:   map[string]*genMapDef{},
	}

	for name, struc := range src.Structs {
		dst.Structs[name] = structWireToGen(struc)
	}

	for name, multimap := range src.Multimaps {
		dst.Multimaps[name] = multimapWireToGen(multimap)
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

func multimapWireToGen(src *schema.Multimap) *genMapDef {
	return &genMapDef{
		Name:  src.Name,
		Key:   multimapFieldWireToAst(src.Key),
		Value: multimapFieldWireToAst(src.Value),
	}
}

func multimapFieldWireToAst(src schema.MultimapField) genMapFieldDef {
	return genMapFieldDef{
		Type: typeWireToGen(src.Type),
		//Recursive: src.Recursive,
	}
}

func structWireToGen(src *schema.Struct) *genStructDef {
	dst := &genStructDef{
		Name:   src.Name,
		OneOf:  src.OneOf,
		IsRoot: src.IsRoot,
		Dict:   src.DictName,
	}

	for i := range src.Fields {
		dst.Fields = append(dst.Fields, structFieldWireToAst(&src.Fields[i]))
	}
	return dst
}

func structFieldWireToAst(src *schema.StructField) *genStructFieldDef {
	dst := &genStructFieldDef{
		Name:     src.Name,
		Optional: src.Optional,
		//Recursive: src.Recursive,
	}

	dst.Type = typeWireToGen(src.FieldType)

	return dst
}

func typeWireToGen(src schema.FieldType) genFieldTypeRef {
	if src.Primitive != nil {
		return &genPrimitiveTypeRef{
			Type: *src.Primitive,
			Dict: src.DictName,
		}
	}

	if src.Array != nil {
		return &genArrayTypeRef{
			ElemType: typeWireToGen(*src.Array),
		}
	}

	if src.Struct != "" {
		return &genStructTypeRef{
			Name: src.Struct,
		}
	}

	if src.MultiMap != "" {
		return &genMultimapTypeRef{
			Name: src.MultiMap,
		}
	}

	panic("unknown field type")
}
