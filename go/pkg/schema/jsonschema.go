package schema

import (
	"fmt"
	"sort"
)

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

func dictName2Idx(mapping *toWireMapping, dictName string) DictIndex {
	if dictName != "" {
		idx, exists := mapping.dicts[dictName]
		if !exists {
			idx = DictIndex(len(mapping.dicts) + 1)
			mapping.dicts[dictName] = idx
		}
		return idx
	}
	return DictNone
}

func (j *JsonStruct) ToWire(mapping *toWireMapping) WireStruct {
	dst := WireStruct{
		Idx:     mapping.structs[j.Name],
		OneOf:   j.OneOf,
		DictIdx: dictName2Idx(mapping, j.DictName),
		IsRoot:  j.IsRoot,
		Fields:  make([]WireStructField, len(j.Fields)),
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

func (j *JsonStructField) ToWire(mapping *toWireMapping) WireStructField {
	dst := WireStructField{
		WireFieldType: *j.JsonFieldType.ToWire(mapping),
		Optional:      j.Optional,
	}
	return dst
}

type JsonFieldType struct {
	Primitive *PrimitiveFieldType `json:"primitive,omitempty"`
	Array     *JsonFieldType      `json:"array,omitempty"`
	Struct    string              `json:"struct,omitempty"`
	MultiMap  string              `json:"multimap,omitempty"`
	DictName  string              `json:"dict,omitempty"`
}

func (j *JsonFieldType) ToWire(mapping *toWireMapping) *WireFieldType {
	dst := &WireFieldType{}
	if j.Primitive != nil {
		dst.Primitive = j.Primitive
	} else if j.Array != nil {
		dst.Array = j.Array.ToWire(mapping)
	} else if j.Struct != "" {
		dst.Struct = mapping.structs[j.Struct]
	} else if j.MultiMap != "" {
		dst.MultiMap = mapping.multimaps[j.MultiMap]
	} else {
		panic("unknown json field type")
	}
	dst.DictIdx = dictName2Idx(mapping, j.DictName)
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

func (m *JsonMultimap) ToWire(mapping *toWireMapping) WireMultimap {
	dst := WireMultimap{
		Idx:   mapping.multimaps[m.Name],
		Key:   m.Key.ToWire(mapping),
		Value: m.Value.ToWire(mapping),
	}
	return dst
}

type toWireMapping struct {
	structs   map[string]StructIndex
	multimaps map[string]MultimapIndex
	dicts     map[string]DictIndex
}

func (j *JsonSchema) ToWire() (*WireSchema, error) {
	dst := &WireSchema{}

	mapping := &toWireMapping{
		structs:   map[string]StructIndex{},
		multimaps: map[string]MultimapIndex{},
		dicts:     map[string]DictIndex{},
	}

	// Sort for deterministic ordering.
	var structs []string
	for name := range j.Structs {
		structs = append(structs, name)
	}
	sort.Strings(structs)

	// Create struct mapping
	dst.structs = make([]WireStruct, len(structs))
	for i, name := range structs {
		mapping.structs[name] = StructIndex(i + 1)
	}

	var multimaps []string
	for name := range j.Multimaps {
		multimaps = append(multimaps, name)
	}
	sort.Strings(structs)

	// Create multimap mapping
	dst.multimaps = make([]WireMultimap, len(multimaps))
	for i, name := range multimaps {
		mapping.multimaps[name] = MultimapIndex(i + 1)
	}

	// Create wire equivalents
	dst.structs = make([]WireStruct, len(structs))
	for i, name := range structs {
		dst.structs[i] = j.Structs[name].ToWire(mapping)
	}
	dst.multimaps = make([]WireMultimap, len(multimaps))
	for i, name := range multimaps {
		dst.multimaps[i] = j.Multimaps[name].ToWire(mapping)
	}

	dst.MainStruct = mapping.structs[j.MainStruct]

	return dst, nil
}

// PrunedForRoot produces a pruned copy of the schema that includes the specified root
// struct and parts of schema reachable from that root. Unreachable parts of the schema
// are excluded.
func (d *JsonSchema) PrunedForRoot(rootStructName string) (*JsonSchema, error) {
	out := JsonSchema{
		Structs:    map[string]*JsonStruct{},
		Multimaps:  map[string]*JsonMultimap{},
		MainStruct: rootStructName,
	}
	if err := d.copyPrunedStruct(rootStructName, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func (d *JsonSchema) copyPrunedFieldType(fieldType *JsonFieldType, dst *JsonSchema) error {
	if fieldType.Struct != "" {
		if err := d.copyPrunedStruct(fieldType.Struct, dst); err != nil {
			return err
		}
	} else if fieldType.MultiMap != "" {
		if err := d.copyPrunedMultiMap(fieldType.MultiMap, dst); err != nil {
			return err
		}
	} else if fieldType.Array != nil {
		if err := d.copyPrunedFieldType(fieldType.Array, dst); err != nil {
			return err
		}
	}
	return nil
}

func (d *JsonSchema) copyPrunedStruct(strucName string, dst *JsonSchema) error {
	if dst.Structs[strucName] != nil {
		// already copied
		return nil
	}

	srcStruc := d.Structs[strucName]
	if srcStruc == nil {
		return fmt.Errorf("no struct named %s found", strucName)
	}

	dstStruc := &JsonStruct{
		Name:     srcStruc.Name,
		OneOf:    srcStruc.OneOf,
		DictName: srcStruc.DictName,
		IsRoot:   srcStruc.IsRoot,
		Fields:   make([]JsonStructField, len(srcStruc.Fields)),
	}
	dst.Structs[strucName] = dstStruc

	for i := range srcStruc.Fields {
		dstStruc.Fields[i] = srcStruc.Fields[i]
		if err := d.copyPrunedFieldType(&dstStruc.Fields[i].JsonFieldType, dst); err != nil {
			return err
		}
	}

	return nil
}

func (d *JsonSchema) copyPrunedMultiMap(multiMapName string, dst *JsonSchema) error {
	if dst.Multimaps[multiMapName] != nil {
		// already copied
		return nil
	}

	srcMultiMap := d.Multimaps[multiMapName]
	if srcMultiMap == nil {
		return fmt.Errorf("no multimap named %s found", multiMapName)
	}

	dstMultimap := &JsonMultimap{
		Name:  srcMultiMap.Name,
		Key:   srcMultiMap.Key,
		Value: srcMultiMap.Value,
	}
	dst.Multimaps[srcMultiMap.Name] = dstMultimap

	if err := d.copyPrunedFieldType(&dstMultimap.Key.Type, dst); err != nil {
		return err
	}

	if err := d.copyPrunedFieldType(&dstMultimap.Value.Type, dst); err != nil {
		return err
	}

	return nil
}
