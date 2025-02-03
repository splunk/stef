package schema

import (
	"fmt"
)

// Schema is a STEF schema description, serializable in JSON format.
type Schema struct {
	PackageName string               `json:"package,omitempty"`
	Structs     map[string]*Struct   `json:"structs"`
	Multimaps   map[string]*Multimap `json:"multimaps"`
	MainStruct  string               `json:"main"`
}

type Compatibility int

const (
	CompatibilityExact Compatibility = iota
	CompatibilitySuperset
	CompatibilityIncompatible
)

// Compatible checks backward compatibility of this schema with oldSchema.
// If the schemas are incompatible returns CompatibilityIncompatible and an error.
func (d *Schema) Compatible(oldSchema *Schema) (Compatibility, error) {
	if d.MainStruct != oldSchema.MainStruct {
		return CompatibilityIncompatible,
			fmt.Errorf(
				"mismatched main structure names (old=%s, new=%s)",
				oldSchema.MainStruct, d.MainStruct,
			)
	}

	// Exact compatibility is only possible if the number of structs is exactly the same.
	exact := len(d.Structs) == len(oldSchema.Structs)

	for name, oldStruc := range oldSchema.Structs {
		newStruc, ok := d.Structs[name]
		if !ok {
			return CompatibilityIncompatible,
				fmt.Errorf("struct %s does not exist in new schema", name)
		}
		comp, err := d.compatibleStruct(name, newStruc, oldStruc)
		if err != nil {
			return CompatibilityIncompatible, err
		}
		if comp == CompatibilitySuperset {
			exact = false
		}
	}

	for name, oldMap := range oldSchema.Multimaps {
		newMap, ok := d.Multimaps[name]
		if !ok {
			return CompatibilityIncompatible,
				fmt.Errorf("multimap %s does not exist in new schema", name)
		}
		comp, err := d.compatibleMultimap(name, newMap, oldMap)
		if err != nil {
			return CompatibilityIncompatible, err
		}
		if comp == CompatibilitySuperset {
			exact = false
		}
	}

	if exact {
		return CompatibilityExact, nil
	}

	return CompatibilitySuperset, nil
}

func (d *Schema) compatibleStruct(
	name string, newStruct *Struct, oldStruc *Struct,
) (Compatibility, error) {
	if len(newStruct.Fields) < len(oldStruc.Fields) {
		return CompatibilityIncompatible, fmt.Errorf("new struct %s has fewer fields than old struct", name)
	}

	if newStruct.OneOf != oldStruc.OneOf {
		return CompatibilityIncompatible, fmt.Errorf("new struct %s has different oneof flag than theold struct", name)
	}

	if newStruct.DictName != oldStruc.DictName {
		return CompatibilityIncompatible, fmt.Errorf(
			"new struct %s dictionary name is %s, old struct dictionary name is %s",
			name, newStruct.DictName, oldStruc.DictName,
		)
	}

	// Exact compatibility is only possible if the number of fields is exactly the same.
	exact := len(newStruct.Fields) == len(oldStruc.Fields)

	for i := range oldStruc.Fields {
		newField := newStruct.Fields[i]
		oldField := oldStruc.Fields[i]
		if err := isCompatibleField(name, i, &newField, &oldField); err != nil {
			return CompatibilityIncompatible, err
		}
	}

	if exact {
		return CompatibilityExact, nil
	}

	return CompatibilitySuperset, nil
}

func (d *Schema) compatibleMultimap(
	name string, newMap *Multimap, oldMap *Multimap,
) (Compatibility, error) {
	if !isCompatibleFieldType(&newMap.Key.Type, &oldMap.Key.Type) {
		return CompatibilityIncompatible,
			fmt.Errorf("multimap %s key type does not match", name)
	}
	if !isCompatibleFieldType(&newMap.Value.Type, &oldMap.Value.Type) {
		return CompatibilityIncompatible,
			fmt.Errorf("multimap %s value type does not match", name)
	}
	return CompatibilityExact, nil
}

func isCompatibleField(
	structName string, fieldIndex int, newField *StructField, oldField *StructField,
) error {
	if newField.Optional != oldField.Optional {
		return fmt.Errorf(
			"field %d in new struct %s has different optional flag than the old struct",
			fieldIndex, structName,
		)
	}

	if !isCompatibleFieldType(&newField.FieldType, &oldField.FieldType) {
		return fmt.Errorf(
			"field %d in new struct %s has a different type than the old struct",
			fieldIndex, structName,
		)
	}

	return nil
}

func isCompatibleFieldType(
	newField *FieldType, oldField *FieldType,
) bool {
	if (newField.Primitive == nil) != (oldField.Primitive == nil) {
		return false
	}

	if newField.Primitive != nil {
		if *newField.Primitive != *oldField.Primitive {
			return false
		}
	}

	if (newField.Array == nil) != (oldField.Array == nil) {
		return false
	}

	if newField.Array != nil {
		if !isCompatibleFieldType(newField.Array, oldField.Array) {
			return false
		}
	}

	if newField.Struct != oldField.Struct {
		return false
	}

	if newField.MultiMap != oldField.MultiMap {
		return false
	}

	if newField.DictName != oldField.DictName {
		return false
	}

	return true
}

// PrunedForRoot produces a pruned copy of the schema that includes the specified root
// struct and parts of schema reachable from that root. Unreachable parts of the schema
// are excluded.
func (d *Schema) PrunedForRoot(rootStructName string) (*Schema, error) {
	out := Schema{
		Structs:    map[string]*Struct{},
		Multimaps:  map[string]*Multimap{},
		MainStruct: rootStructName,
	}
	if err := d.copyPrunedStruct(rootStructName, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func (d *Schema) copyPrunedFieldType(fieldType *FieldType, dst *Schema) error {
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

func (d *Schema) copyPrunedStruct(strucName string, dst *Schema) error {
	if dst.Structs[strucName] != nil {
		// already copied
		return nil
	}

	srcStruc := d.Structs[strucName]
	if srcStruc == nil {
		return fmt.Errorf("no struct named %s found", strucName)
	}

	dstStruc := &Struct{
		Name:     strucName,
		OneOf:    srcStruc.OneOf,
		DictName: srcStruc.DictName,
		IsRoot:   srcStruc.IsRoot,
		Fields:   make([]StructField, len(srcStruc.Fields)),
	}
	dst.Structs[strucName] = dstStruc

	for i := range srcStruc.Fields {
		dstStruc.Fields[i] = srcStruc.Fields[i]
		if err := d.copyPrunedFieldType(&dstStruc.Fields[i].FieldType, dst); err != nil {
			return err
		}
	}

	return nil
}

func (d *Schema) copyPrunedMultiMap(multiMapName string, dst *Schema) error {
	if dst.Multimaps[multiMapName] != nil {
		// already copied
		return nil
	}

	srcMultiMap := d.Multimaps[multiMapName]
	if srcMultiMap == nil {
		return fmt.Errorf("no multimap named %s found", multiMapName)
	}

	dstMultimap := &Multimap{
		Name:  multiMapName,
		Key:   srcMultiMap.Key,
		Value: srcMultiMap.Value,
	}
	dst.Multimaps[multiMapName] = dstMultimap

	if err := d.copyPrunedFieldType(&dstMultimap.Key.Type, dst); err != nil {
		return err
	}

	if err := d.copyPrunedFieldType(&dstMultimap.Value.Type, dst); err != nil {
		return err
	}

	return nil
}

func (d *Schema) ToWire() WireSchema {
	w := WireSchema{
		StructFieldCount: make(map[string]uint),
	}
	for k, v := range d.Structs {
		w.StructFieldCount[k] = uint(len(v.Fields))
	}
	return w
}

type Struct struct {
	Name     string        `json:"name,omitempty"`
	OneOf    bool          `json:"oneof,omitempty"`
	DictName string        `json:"dict,omitempty"`
	IsRoot   bool          `json:"root,omitempty"`
	Fields   []StructField `json:"fields"`
}

type StructField struct {
	FieldType
	Name     string `json:"name,omitempty"`
	Optional bool   `json:"optional,omitempty"`
}

type PrimitiveFieldType int

const (
	PrimitiveTypeInt64 PrimitiveFieldType = iota
	PrimitiveTypeUint64
	PrimitiveTypeFloat64
	PrimitiveTypeBool
	PrimitiveTypeString
	PrimitiveTypeBytes
)

type FieldType struct {
	Primitive *PrimitiveFieldType `json:"primitive,omitempty"`
	Array     *FieldType          `json:"array,omitempty"`
	Struct    string              `json:"struct,omitempty"`
	MultiMap  string              `json:"multimap,omitempty"`
	DictName  string              `json:"dict,omitempty"`
}

type MultimapField struct {
	Type FieldType `json:"type"`
}

type Multimap struct {
	Name  string        `json:"name,omitempty"`
	Key   MultimapField `json:"key"`
	Value MultimapField `json:"value"`
}
