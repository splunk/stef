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

	return dst, nil
}

func (s *genSchema) resolveType(typ genFieldTypeRef) error {
	if ref, ok := typ.(*genStructTypeRef); ok {
		ref.Def = s.Structs[ref.Name]
		if ref.Def == nil {
			return fmt.Errorf("struct %s not found", ref.Name)
		}
	}

	if ref, ok := typ.(*genArrayTypeRef); ok {
		if err := s.resolveType(ref.ElemType); err != nil {
			return err
		}
	}

	if ref, ok := typ.(*genMultimapTypeRef); ok {
		ref.Def = s.Multimaps[ref.Name]
		if ref.Def == nil {
			return fmt.Errorf("multimap %s not found", ref.Name)
		}
	}

	ref, ok := typ.(*genPrimitiveTypeRef)
	if ok && ref.Enum != "" {
		ref.EnumDef = s.Enums[ref.Enum]
		if ref.EnumDef == nil {
			return fmt.Errorf("enum %s not found", ref.Enum)
		}
	}

	return nil
}

func (s *genSchema) resolveRefs() error {
	for _, struc := range s.Structs {
		for _, field := range struc.Fields {
			if err := s.resolveType(field.Type); err != nil {
				return err
			}
		}
	}
	for _, mp := range s.Multimaps {
		if err := s.resolveType(mp.Key.Type); err != nil {
			return err
		}
		if err := s.resolveType(mp.Value.Type); err != nil {
			return err
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
		genFieldDef{
			Type: typeWireToGen(src.Type, lang),
		},
	}
}

func structWireToGen(src *schema.Struct, lang Lang) *genStructDef {
	dst := &genStructDef{
		Def:    src,
		Name:   src.Name,
		OneOf:  src.OneOf,
		IsRoot: src.IsRoot,
		Dict:   src.DictName,
	}

	for i := range src.Fields {
		dst.Fields = append(dst.Fields, structFieldWireToAst(src.Fields[i], lang))
	}
	return dst
}

func structFieldWireToAst(src *schema.StructField, lang Lang) *genStructFieldDef {
	dst := &genStructFieldDef{
		genFieldDef: genFieldDef{Recursive: src.Recursive()},
		Name:        src.Name,
		Optional:    src.Optional,
	}

	dst.Type = typeWireToGen(src.FieldType, lang)

	return dst
}

func typeWireToGen(src schema.FieldType, lang Lang) genFieldTypeRef {
	if src.Primitive != nil {
		return &genPrimitiveTypeRef{
			Lang: lang,
			Type: src.Primitive.Type,
			Dict: src.DictName,
			Enum: src.Enum,
		}
	}

	if src.Array != nil {
		return &genArrayTypeRef{
			Lang:     lang,
			ElemType: typeWireToGen(src.Array.ElemType, lang),
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
