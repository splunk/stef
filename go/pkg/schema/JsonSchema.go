package schema

import "sort"

// JsonSchema is a STEF schema description.
type JsonSchema struct {
	PackageName string                   `json:"package,omitempty"`
	Structs     map[string]*JsonStruct   `json:"structs"`
	Multimaps   map[string]*JsonMultimap `json:"multimaps"`
	MainStruct  string                   `json:"main"`
}

type JsonStruct struct {
	Name     string            `json:"name,omitempty"`
	OneOf    bool              `json:"oneof,omitempty"`
	DictName string            `json:"dict,omitempty"`
	IsRoot   bool              `json:"root,omitempty"`
	Fields   []JsonStructField `json:"fields"`
}

func (j *JsonStruct) ToWire(mapping *toWireMapping) Struct {
	dst := Struct{
		Name:     mapping.structs[j.Name],
		OneOf:    j.OneOf,
		DictName: j.DictName,
		IsRoot:   j.IsRoot,
		Fields:   make([]StructField, len(j.Fields)),
	}
	for i := range j.Fields {
		dst.Fields[i] = j.Fields[i].ToWire(mapping)
	}
	return dst
}

type JsonStructField struct {
	JsonFieldType
	Name     string `json:"name,omitempty"`
	Optional bool   `json:"optional,omitempty"`
}

func (j *JsonStructField) ToWire(mapping *toWireMapping) StructField {
	dst := StructField{
		FieldType: *j.JsonFieldType.ToWire(mapping),
		Optional:  j.Optional,
	}
	return dst
}

type JsonFieldType struct {
	Primitive  *PrimitiveFieldType `json:"primitive,omitempty"`
	Array      *JsonFieldType      `json:"array,omitempty"`
	JsonStruct string              `json:"struct,omitempty"`
	MultiMap   string              `json:"multimap,omitempty"`
	DictName   string              `json:"dict,omitempty"`
}

func (j *JsonFieldType) ToWire(mapping *toWireMapping) *FieldType {
	dst := &FieldType{}
	if j.Primitive != nil {
		dst.Primitive = j.Primitive
	} else if j.Array != nil {
		dst.Array = j.Array.ToWire(mapping)
	} else if j.JsonStruct != "" {
		dst.Struct = mapping.structs[j.JsonStruct]
	} else if j.MultiMap != "" {
		dst.MultiMap = mapping.multimaps[j.MultiMap]
	} else {
		panic("unknown json field type")
	}
	return dst
}

type JsonMultimapField struct {
	Type JsonFieldType `json:"type"`
}

func (f *JsonMultimapField) ToWire(mapping *toWireMapping) MultimapField {
	return MultimapField{
		Type: *f.Type.ToWire(mapping),
	}
}

type JsonMultimap struct {
	Name  string            `json:"name,omitempty"`
	Key   JsonMultimapField `json:"key"`
	Value JsonMultimapField `json:"value"`
}

func (m *JsonMultimap) ToWire(mapping *toWireMapping) Multimap {
	dst := Multimap{
		Name:  mapping.multimaps[m.Name],
		Key:   m.Key.ToWire(mapping),
		Value: m.Value.ToWire(mapping),
	}
	return dst
}

type toWireMapping struct {
	structs   map[string]StructIndex
	multimaps map[string]MultimapIndex
}

func (j *JsonSchema) ToWire() (*Schema, error) {
	dst := &Schema{}

	mapping := &toWireMapping{
		structs:   map[string]StructIndex{},
		multimaps: map[string]MultimapIndex{},
	}

	// Sort for deterministic ordering.
	var structs []string
	for name := range j.Structs {
		structs = append(structs, name)
	}
	sort.Strings(structs)

	// Create struct mapping
	dst.Structs = make([]Struct, len(structs))
	for i, name := range structs {
		mapping.structs[name] = StructIndex(i + 1)
	}

	var multimaps []string
	for name := range j.Multimaps {
		multimaps = append(multimaps, name)
	}
	sort.Strings(structs)

	// Create multimap mapping
	dst.Multimaps = make([]Multimap, len(multimaps))
	for i, name := range multimaps {
		mapping.multimaps[name] = MultimapIndex(i + 1)
	}

	// Create wire equivalents
	dst.Structs = make([]Struct, len(structs))
	for i, name := range structs {
		dst.Structs[i+1] = j.Structs[name].ToWire(mapping)
	}
	dst.Multimaps = make([]Multimap, len(multimaps))
	for i, name := range multimaps {
		dst.Multimaps[i+1] = j.Multimaps[name].ToWire(mapping)
	}

	dst.MainStruct = mapping.structs[j.MainStruct]

	return dst, nil
}
