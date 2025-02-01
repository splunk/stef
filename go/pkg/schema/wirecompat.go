package schema

import "fmt"

type Compatibility int

const (
	CompatibilityExact Compatibility = iota
	CompatibilitySuperset
	CompatibilityIncompatible
)

type compatMapping struct {
	// Struct mapping. Key is old index, value is new index.
	structIdxs map[StructIndex]StructIndex

	// Multimap mapping. Key is old index, value is new index.
	multimapIdxs map[MultimapIndex]MultimapIndex

	// Dict mapping. Key is old index, value is new index.
	dictIdxs map[DictIndex]DictIndex
}

func (d *WireSchema) Struct(idx StructIndex) WireStruct {
	return d.structs[idx-1]
}

func (d *WireSchema) Multimap(idx MultimapIndex) WireMultimap {
	return d.multimaps[idx-1]
}

func (m *compatMapping) traverse(old *WireSchema, new *WireSchema) {
	m.structIdxs[old.MainStruct] = new.MainStruct
	m.traverseStruct(old, new, old.MainStruct)
}

func (m *compatMapping) traverseStruct(old *WireSchema, new *WireSchema, oldStructIdx StructIndex) {
	oldStr := old.Struct(oldStructIdx)
	newIdx, exists := m.structIdxs[oldStructIdx]
	if !exists {
		return
	}
	newStr := new.Struct(newIdx)

	if oldStr.DictIdx != DictNone {
		m.dictIdxs[oldStr.DictIdx] = newStr.DictIdx
	}

	fieldCount := min(len(oldStr.Fields), len(newStr.Fields))
	for i := 0; i < fieldCount; i++ {
		oldField := oldStr.Fields[i]
		m.traverseField(old, new, &oldField.WireFieldType, &newStr.Fields[i].WireFieldType)
	}
}

func (m *compatMapping) traverseMultimap(old *WireSchema, new *WireSchema, oldMultiMapIdx MultimapIndex) {
	oldMultimap := old.Multimap(oldMultiMapIdx)
	newIdx, exists := m.multimapIdxs[oldMultiMapIdx]
	if !exists {
		return
	}
	newMultimap := new.Multimap(newIdx)

	m.traverseField(old, new, &oldMultimap.Key.Type, &newMultimap.Key.Type)
	m.traverseField(old, new, &oldMultimap.Value.Type, &newMultimap.Value.Type)
}

func (m *compatMapping) traverseField(old *WireSchema, new *WireSchema, oldField, newField *WireFieldType) {
	if oldField.Struct != 0 {
		if _, exists := m.structIdxs[oldField.Struct]; !exists {
			m.structIdxs[oldField.Struct] = newField.Struct
			m.traverseStruct(old, new, oldField.Struct)
		}
	} else if oldField.MultiMap != 0 {
		if _, exists := m.multimapIdxs[oldField.MultiMap]; !exists {
			m.multimapIdxs[oldField.MultiMap] = newField.MultiMap
			m.traverseMultimap(old, new, oldField.MultiMap)
		}
	} else if oldField.Array != nil {
		m.traverseField(old, new, oldField.Array, oldField.Array)
	}

	if oldField.DictIdx != DictNone {
		m.dictIdxs[oldField.DictIdx] = newField.DictIdx
	}
}

// Compatible checks backward compatibility of this schema with oldSchema.
// If the schemas are incompatible returns CompatibilityIncompatible and an error.
func (d *WireSchema) Compatible(oldSchema *WireSchema) (Compatibility, error) {

	compat := compatMapping{
		structIdxs:   map[StructIndex]StructIndex{},
		multimapIdxs: map[MultimapIndex]MultimapIndex{},
		dictIdxs:     map[DictIndex]DictIndex{},
	}
	compat.structIdxs[oldSchema.MainStruct] = d.MainStruct

	compat.traverse(oldSchema, d)

	// Exact compatibility is only possible if the number of structs is exactly the same.
	exact := len(d.structs) == len(oldSchema.structs)

	for oldIdx, newIdx := range compat.structIdxs {
		oldStruc := oldSchema.Struct(oldIdx)
		//if !ok {
		//	panic("compat struct is invalid")
		//}
		newStruc := d.Struct(newIdx)
		//if !ok {
		//	return CompatibilityIncompatible,
		//		fmt.Errorf(
		//			"new struct %s is expected to correspond to old struct %s, but does not exist in new schema",
		//			newIdx, oldIdx,
		//		)
		//}
		comp, err := d.compatibleStruct(compat, &newStruc, &oldStruc)
		if err != nil {
			return CompatibilityIncompatible, err
		}
		if comp == CompatibilitySuperset {
			exact = false
		}
	}

	for oldIdx, newIdx := range compat.multimapIdxs {
		oldMap := oldSchema.Multimap(oldIdx)
		//if !ok {
		//	panic("compat struct is invalid")
		//}
		newMap := d.Multimap(newIdx)
		//if !ok {
		//	return CompatibilityIncompatible,
		//		fmt.Errorf(
		//			"new multimap %s is expected to correspond to old multimap %s, but does not exist in new schema",
		//			newIdx, oldIdx,
		//		)
		//}
		comp, err := d.compatibleMultimap(compat, oldIdx, &newMap, &oldMap)
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

func (d *WireSchema) compatibleStruct(
	compat compatMapping,
	newStruct *WireStruct, oldStruct *WireStruct,
) (Compatibility, error) {
	if len(newStruct.Fields) < len(oldStruct.Fields) {
		return CompatibilityIncompatible, fmt.Errorf(
			"new struct %v has fewer fields than old struct %v",
			newStruct.Idx, oldStruct.Idx,
		)
	}

	if newStruct.OneOf != oldStruct.OneOf {
		return CompatibilityIncompatible, fmt.Errorf(
			"new struct %v has different oneof flag than the old struct %v",
			newStruct.Idx, oldStruct.Idx,
		)
	}

	if newStruct.DictIdx != compat.dictIdxs[oldStruct.DictIdx] {
		return CompatibilityIncompatible, fmt.Errorf(
			"new struct %v dictionary is %v, old struct %v dictionary is %v",
			newStruct.Idx, newStruct.DictIdx, oldStruct.Idx, oldStruct.DictIdx,
		)
	}

	// Exact compatibility is only possible if the number of fields is exactly the same.
	exact := len(newStruct.Fields) == len(oldStruct.Fields)

	for i := range oldStruct.Fields {
		newField := &newStruct.Fields[i]
		oldField := &oldStruct.Fields[i]
		if err := isCompatibleField(compat, oldStruct.Idx, i, newField, oldField); err != nil {
			return CompatibilityIncompatible, err
		}
	}

	if exact {
		return CompatibilityExact, nil
	}

	return CompatibilitySuperset, nil
}

func (d *WireSchema) compatibleMultimap(
	compat compatMapping,
	name MultimapIndex, newMap *WireMultimap, oldMap *WireMultimap,
) (Compatibility, error) {
	if !isCompatibleFieldType(compat, &newMap.Key.Type, &oldMap.Key.Type) {
		return CompatibilityIncompatible,
			fmt.Errorf("multimap %v key type does not match", name)
	}
	if !isCompatibleFieldType(compat, &newMap.Value.Type, &oldMap.Value.Type) {
		return CompatibilityIncompatible,
			fmt.Errorf("multimap %v value type does not match", name)
	}
	return CompatibilityExact, nil
}

func isCompatibleField(
	compat compatMapping,
	oldStructIdx StructIndex, fieldIndex int, newField *WireStructField, oldField *WireStructField,
) error {
	if newField.Optional != oldField.Optional {
		return fmt.Errorf(
			"field %d in new struct %v has different optional flag than in the old struct %v",
			fieldIndex, compat.structIdxs[oldStructIdx], oldStructIdx,
		)
	}

	if !isCompatibleFieldType(compat, &newField.WireFieldType, &oldField.WireFieldType) {
		return fmt.Errorf(
			"field %d in new struct %v has a different type than in the old struct %v",
			fieldIndex, compat.structIdxs[oldStructIdx], oldStructIdx,
		)
	}

	return nil
}

func isCompatibleFieldType(
	compat compatMapping,
	newField *WireFieldType, oldField *WireFieldType,
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
		if !isCompatibleFieldType(compat, newField.Array, oldField.Array) {
			return false
		}
	}

	if newField.Struct != compat.structIdxs[oldField.Struct] {
		return false
	}

	if newField.MultiMap != compat.multimapIdxs[oldField.MultiMap] {
		return false
	}

	if newField.DictIdx != compat.dictIdxs[oldField.DictIdx] {
		return false
	}

	return true
}
