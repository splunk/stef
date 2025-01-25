package schema

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"

	"github.com/splunk/stef/go/pkg/internal"
)

// Schema is a STEF schema description.
type Schema struct {
	PackageName string               `json:"package,omitempty"`
	Structs     map[string]*Struct   `json:"structs"`
	Multimaps   map[string]*Multimap `json:"multimaps"`
	MainStruct  string               `json:"main"`
}

const (
	MaxStructOrMultimapCount = 256
	MaxStructFieldCount      = 256
)

var (
	errStructOrMultimapCountLimit = errors.New("struct or multimap count limit exceeded")
	errStructFieldCountLimit      = errors.New("struct field count limit exceeded")
)

type Compatibility int

const (
	CompatibilityExact Compatibility = iota
	CompatibilitySuperset
	CompatibilityIncompatible
)

type compatMapping struct {
	// Struct name mapping. Key is old name, value is new name.
	structNames map[string]string

	// Multimap name mapping. Key is old name, value is new name.
	multimapNames map[string]string
}

func (m *compatMapping) traverse(old *Schema, new *Schema) {
	m.structNames[old.MainStruct] = new.MainStruct
	m.traverseStruct(old, new, old.MainStruct)
}

func (m *compatMapping) traverseStruct(old *Schema, new *Schema, oldStructName string) {
	oldStr := old.Structs[oldStructName]
	newStr := new.Structs[m.structNames[oldStructName]]
	if newStr == nil {
		return
	}

	fieldCount := min(len(oldStr.Fields), len(newStr.Fields))
	for i := 0; i < fieldCount; i++ {
		oldField := oldStr.Fields[i]
		m.traverseField(old, new, &oldField.FieldType, &newStr.Fields[i].FieldType)
	}
}

func (m *compatMapping) traverseMultimap(old *Schema, new *Schema, oldMultiMapName string) {
	oldMultimap := old.Multimaps[oldMultiMapName]
	newMultimap := new.Multimaps[m.multimapNames[oldMultiMapName]]
	if newMultimap == nil {
		return
	}

	m.traverseField(old, new, &oldMultimap.Key.Type, &newMultimap.Key.Type)
	m.traverseField(old, new, &oldMultimap.Value.Type, &newMultimap.Value.Type)
}

func (m *compatMapping) traverseField(old *Schema, new *Schema, oldField, newField *FieldType) {
	if oldField.Struct != "" {
		if _, exists := m.structNames[oldField.Struct]; !exists {
			m.structNames[oldField.Struct] = newField.Struct
			m.traverseStruct(old, new, oldField.Struct)
		}
	} else if oldField.MultiMap != "" {
		if _, exists := m.multimapNames[oldField.MultiMap]; !exists {
			m.multimapNames[oldField.MultiMap] = newField.MultiMap
			m.traverseMultimap(old, new, oldField.MultiMap)
		}
	} else if oldField.Array != nil {
		m.traverseField(old, new, oldField.Array, oldField.Array)
	}
}

// Compatible checks backward compatibility of this schema with oldSchema.
// If the schemas are incompatible returns CompatibilityIncompatible and an error.
func (d *Schema) Compatible(oldSchema *Schema) (Compatibility, error) {

	compat := compatMapping{
		structNames:   map[string]string{},
		multimapNames: map[string]string{},
	}
	compat.structNames[oldSchema.MainStruct] = d.MainStruct

	compat.traverse(oldSchema, d)

	// Exact compatibility is only possible if the number of structs is exactly the same.
	exact := len(d.Structs) == len(oldSchema.Structs)

	for oldName, newName := range compat.structNames {
		oldStruc, ok := oldSchema.Structs[oldName]
		if !ok {
			panic("compat struct is invalid")
		}
		newStruc, ok := d.Structs[newName]
		if !ok {
			return CompatibilityIncompatible,
				fmt.Errorf(
					"new struct %s is expected to correspond to old struct %s, but does not exist in new schema",
					newName, oldName,
				)
		}
		comp, err := d.compatibleStruct(compat, newStruc, oldStruc)
		if err != nil {
			return CompatibilityIncompatible, err
		}
		if comp == CompatibilitySuperset {
			exact = false
		}
	}

	for oldName, newName := range compat.multimapNames {
		oldMap, ok := oldSchema.Multimaps[oldName]
		if !ok {
			panic("compat struct is invalid")
		}
		newMap, ok := d.Multimaps[newName]
		if !ok {
			return CompatibilityIncompatible,
				fmt.Errorf(
					"new multimap %s is expected to correspond to old multimap %s, but does not exist in new schema",
					newName, oldName,
				)
		}
		comp, err := d.compatibleMultimap(compat, oldName, newMap, oldMap)
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
	compat compatMapping,
	newStruct *Struct, oldStruct *Struct,
) (Compatibility, error) {
	if len(newStruct.Fields) < len(oldStruct.Fields) {
		return CompatibilityIncompatible, fmt.Errorf(
			"new struct %s has fewer fields than old struct %s",
			newStruct.Name, oldStruct.Name,
		)
	}

	if newStruct.OneOf != oldStruct.OneOf {
		return CompatibilityIncompatible, fmt.Errorf(
			"new struct %s has different oneof flag than the old struct %s",
			newStruct.Name, oldStruct.Name,
		)
	}

	if newStruct.DictName != oldStruct.DictName {
		return CompatibilityIncompatible, fmt.Errorf(
			"new struct %s dictionary name is %s, old struct %s dictionary name is %s",
			newStruct.Name, newStruct.DictName, oldStruct.Name, oldStruct.DictName,
		)
	}

	// Exact compatibility is only possible if the number of fields is exactly the same.
	exact := len(newStruct.Fields) == len(oldStruct.Fields)

	for i := range oldStruct.Fields {
		newField := &newStruct.Fields[i]
		oldField := &oldStruct.Fields[i]
		if err := isCompatibleField(compat, oldStruct.Name, i, newField, oldField); err != nil {
			return CompatibilityIncompatible, err
		}
	}

	if exact {
		return CompatibilityExact, nil
	}

	return CompatibilitySuperset, nil
}

func (d *Schema) compatibleMultimap(
	compat compatMapping,
	name string, newMap *Multimap, oldMap *Multimap,
) (Compatibility, error) {
	if !isCompatibleFieldType(compat, &newMap.Key.Type, &oldMap.Key.Type) {
		return CompatibilityIncompatible,
			fmt.Errorf("multimap %s key type does not match", name)
	}
	if !isCompatibleFieldType(compat, &newMap.Value.Type, &oldMap.Value.Type) {
		return CompatibilityIncompatible,
			fmt.Errorf("multimap %s value type does not match", name)
	}
	return CompatibilityExact, nil
}

func isCompatibleField(
	compat compatMapping,
	oldStructName string, fieldIndex int, newField *StructField, oldField *StructField,
) error {
	if newField.Optional != oldField.Optional {
		return fmt.Errorf(
			"field %d in new struct %s has different optional flag than in the old struct %s",
			fieldIndex, compat.structNames[oldStructName], oldStructName,
		)
	}

	if !isCompatibleFieldType(compat, &newField.FieldType, &oldField.FieldType) {
		return fmt.Errorf(
			"field %d in new struct %s has a different type than in the old struct %s",
			fieldIndex, compat.structNames[oldStructName], oldStructName,
		)
	}

	return nil
}

func isCompatibleFieldType(
	compat compatMapping,
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
		if !isCompatibleFieldType(compat, newField.Array, oldField.Array) {
			return false
		}
	}

	if newField.Struct != compat.structNames[oldField.Struct] {
		return false
	}

	if newField.MultiMap != compat.multimapNames[oldField.MultiMap] {
		return false
	}

	if newField.DictName != oldField.DictName {
		return false
	}

	return true
}

// Minify removes data that is not necessary for wire format identification (such as field names).
// Typically, Minify is used before the schema is serialized and sent over network
// to avoid unnecessary overhead.
func (d *Schema) Minify() {
	d.PackageName = ""
	for _, struc := range d.Structs {
		struc.minify()
	}

	for _, m := range d.Multimaps {
		m.minify()
	}
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

/*
Binary serialization format:

Schema {
	MainStruct:    String
	StructCount:   U64
	*Struct:       Struct
	MultimapCount: U64
	*Multimap:     Multimap
}
*/

// Serialize the schema to binary format.
func (d *Schema) Serialize(dst *bytes.Buffer) error {
	if err := internal.WriteString(d.MainStruct, dst); err != nil {
		return nil
	}

	if err := internal.WriteUvarint(uint64(len(d.Structs)), dst); err != nil {
		return err
	}

	// Sort for deterministic serialization.
	var structs []string
	for name := range d.Structs {
		structs = append(structs, name)
	}
	sort.Strings(structs)

	for _, name := range structs {
		str := d.Structs[name]
		if err := str.serialize(name, dst); err != nil {
			return nil
		}
	}

	if err := internal.WriteUvarint(uint64(len(d.Multimaps)), dst); err != nil {
		return err
	}

	var multimaps []string
	for name := range d.Multimaps {
		multimaps = append(multimaps, name)
	}
	sort.Strings(structs)

	for _, name := range multimaps {
		mm := d.Multimaps[name]
		if err := mm.serialize(name, dst); err != nil {
			return nil
		}
	}

	return nil
}

// Deserialize the schema from binary format.
func (d *Schema) Deserialize(src *bytes.Buffer) error {
	var err error
	d.MainStruct, err = internal.ReadString(src)
	if err != nil {
		return err
	}

	count, err := binary.ReadUvarint(src)
	if err != nil {
		return err
	}

	if count > MaxStructOrMultimapCount {
		return errStructOrMultimapCountLimit
	}

	d.Structs = make(map[string]*Struct, count)
	for i := 0; i < int(count); i++ {
		var str Struct
		if err := str.deserialize(src); err != nil {
			return err
		}
		d.Structs[str.Name] = &str
		//str.Name = ""
	}

	count, err = binary.ReadUvarint(src)
	if err != nil {
		return err
	}

	if count > MaxStructOrMultimapCount {
		return errStructOrMultimapCountLimit
	}

	d.Multimaps = make(map[string]*Multimap, count)
	for i := 0; i < int(count); i++ {
		var mm Multimap
		if err := mm.deserialize(src); err != nil {
			return err
		}
		d.Multimaps[mm.Name] = &mm
		//mm.Name = ""
	}

	return nil
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
		Name:     srcStruc.Name,
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

type Struct struct {
	Name     string        `json:"name,omitempty"`
	OneOf    bool          `json:"oneof,omitempty"`
	DictName string        `json:"dict,omitempty"`
	IsRoot   bool          `json:"root,omitempty"`
	Fields   []StructField `json:"fields"`
}

func (s *Struct) minify() {
	// Name is not needed to identify wire format. It is already
	// recording as the containing map element key.
	//s.Name = ""

	for i := range s.Fields {
		s.Fields[i].minify()
	}
}

/*
Binary serialization format:

Struct {
	Flag: 8
	Name: String
	/DictName: String/
	FieldCount: U64
	*Field: FieldType
}
*/

type structFlag byte

const (
	structFlagIsRoot structFlag = 1 << iota
	structFlagOneOf
	structFlagHasDict
)

func (s *Struct) serialize(name string, dst *bytes.Buffer) error {
	var flags structFlag
	if s.IsRoot {
		flags |= structFlagIsRoot
	}
	if s.OneOf {
		flags |= structFlagOneOf
	}
	if s.DictName != "" {
		flags |= structFlagHasDict
	}
	if err := dst.WriteByte(byte(flags)); err != nil {
		return err
	}
	if err := internal.WriteString(name, dst); err != nil {
		return err
	}
	if s.DictName != "" {
		if err := internal.WriteString(s.DictName, dst); err != nil {
			return err
		}
	}

	if err := internal.WriteUvarint(uint64(len(s.Fields)), dst); err != nil {
		return err
	}

	for _, field := range s.Fields {
		if err := field.serialize(dst); err != nil {
			return err
		}
	}

	return nil
}

func (s *Struct) deserialize(buf *bytes.Buffer) error {
	f, err := buf.ReadByte()
	if err != nil {
		return err
	}
	flags := structFlag(f)

	if flags&structFlagIsRoot != 0 {
		s.IsRoot = true
	}
	if flags&structFlagOneOf != 0 {
		s.OneOf = true
	}

	s.Name, err = internal.ReadString(buf)
	if err != nil {
		return err
	}

	if flags&structFlagHasDict != 0 {
		s.DictName, err = internal.ReadString(buf)
		if err != nil {
			return err
		}
	}

	count, err := binary.ReadUvarint(buf)
	if err != nil {
		return err
	}

	if count > MaxStructFieldCount {
		return errStructFieldCountLimit
	}

	s.Fields = make([]StructField, count)
	for i := range s.Fields {
		if err := s.Fields[i].deserialize(buf); err != nil {
			return err
		}
	}

	return nil
}

type StructField struct {
	FieldType
	Name     string `json:"name,omitempty"`
	Optional bool   `json:"optional,omitempty"`
}

func (f *StructField) minify() {
	// Name is not needed to identify wire format.
	f.Name = ""
}

func (f *StructField) serialize(buf *bytes.Buffer) error {
	return f.FieldType.serialize(buf, f.Optional)
}

func (f *StructField) deserialize(buf *bytes.Buffer) error {
	var err error
	f.Optional, err = f.FieldType.deserialize(buf)
	return err
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

type TypeDescr byte

const (
	TypeDescrArray TypeDescr = TypeDescr(iota + 10)
	TypeDescrStruct
	TypeDescrMultimap
	TypeDescrTypeMask = 0b001111
	TypeDescrOptional = 0b010000
	TypeDescrHasDict  = 0b100000
)

/*
Binary serialization format:

FieldType {
	TypeDescr: 8
	/ElemType: FieldType/
	/StructName: String/
	/MultimapName: String/
	/DictName: String/
}
*/

func (f *FieldType) serialize(buf *bytes.Buffer, optional bool) error {
	var typeDescr TypeDescr

	// Bits 0-3 - type: 0-5 = primitive type
	//                  6-9 = reserved
	//                  10 = array
	//                  11 = struct
	//                  12 = multimap
	//                  13-15 = reserved
	// Bit 4 - optional
	// Bit 5 - has dict.
	// But 6-7 - reserved.

	if f.Primitive != nil {
		typeDescr = TypeDescr(*f.Primitive)
	} else if f.Array != nil {
		typeDescr = TypeDescrArray
	} else if f.Struct != "" {
		typeDescr = TypeDescrStruct
	} else if f.MultiMap != "" {
		typeDescr = TypeDescrMultimap
	} else {
		panic("unknown field type")
	}

	if optional {
		typeDescr |= TypeDescrOptional
	}

	if f.DictName != "" {
		typeDescr |= TypeDescrHasDict
	}

	if err := buf.WriteByte(byte(typeDescr)); err != nil {
		return err
	}

	if f.Array != nil {
		return f.Array.serialize(buf, false)
	} else if f.Struct != "" {
		if err := internal.WriteString(f.Struct, buf); err != nil {
			return err
		}
	} else if f.MultiMap != "" {
		if err := internal.WriteString(f.MultiMap, buf); err != nil {
			return err
		}
	}

	if f.DictName != "" {
		if err := internal.WriteString(f.DictName, buf); err != nil {
			return err
		}
	}

	return nil
}

func (f *FieldType) deserialize(buf *bytes.Buffer) (optional bool, err error) {
	td, err := buf.ReadByte()
	if err != nil {
		return false, err
	}
	typeDescr := TypeDescr(td)

	optional = typeDescr&TypeDescrOptional != 0
	typ := typeDescr & TypeDescrTypeMask
	switch typ {
	case TypeDescrArray:
		var elemType FieldType
		_, err = elemType.deserialize(buf)
		if err != nil {
			return false, err
		}
		f.Array = &elemType

	case TypeDescrStruct:
		f.Struct, err = internal.ReadString(buf)
		if err != nil {
			return false, err
		}

	case TypeDescrMultimap:
		f.MultiMap, err = internal.ReadString(buf)
		if err != nil {
			return false, err
		}

	default:
		if byte(typ) <= byte(PrimitiveTypeBytes) {
			p := PrimitiveFieldType(typ)
			f.Primitive = &p
		} else {
			return false, errors.New("unknown type")
		}
	}

	if typeDescr&TypeDescrHasDict != 0 {
		f.DictName, err = internal.ReadString(buf)
		if err != nil {
			return false, err
		}
	}

	return optional, nil
}

type MultimapField struct {
	Type FieldType `json:"type"`
}

func (f *MultimapField) minify() {
}

func (f *MultimapField) serialize(buf *bytes.Buffer) error {
	return f.Type.serialize(buf, false)
}

func (f *MultimapField) deserialize(buf *bytes.Buffer) error {
	_, err := f.Type.deserialize(buf)
	return err
}

type Multimap struct {
	Name  string        `json:"name,omitempty"`
	Key   MultimapField `json:"key"`
	Value MultimapField `json:"value"`
}

func (m *Multimap) minify() {
	//m.Name = ""
	m.Key.minify()
	m.Value.minify()
}

/*
Binary serialization format:

Multimap {
	Name: String
	Key: FieldType
	Value: FieldType
}
*/

func (m *Multimap) serialize(name string, buf *bytes.Buffer) error {
	if err := internal.WriteString(name, buf); err != nil {
		return err
	}
	if err := m.Key.serialize(buf); err != nil {
		return err
	}
	return m.Value.serialize(buf)
}

func (m *Multimap) deserialize(buf *bytes.Buffer) error {
	var err error
	m.Name, err = internal.ReadString(buf)
	if err != nil {
		return err
	}

	if err := m.Key.deserialize(buf); err != nil {
		return err
	}
	return m.Value.deserialize(buf)
}
