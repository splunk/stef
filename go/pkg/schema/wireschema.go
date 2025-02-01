package schema

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/splunk/stef/go/pkg/internal"
)

type StructIndex uint
type MultimapIndex uint
type DictIndex uint

const (
	StructNone   StructIndex   = 0
	MultimapNone MultimapIndex = 0
	DictNone     DictIndex     = 0
)

// WireSchema is a STEF schema description.
type WireSchema struct {
	structs    []WireStruct
	multimaps  []WireMultimap
	MainStruct StructIndex
}

const (
	MaxStructOrMultimapCount = 256
	MaxStructFieldCount      = 256
)

var (
	errStructOrMultimapCountLimit = errors.New("struct or multimap count limit exceeded")
	errStructFieldCountLimit      = errors.New("struct field count limit exceeded")
)

/*
Binary serialization format:

WireSchema {
	MainStruct:    String
	StructCount:   U64
	*WireStruct:   WireStruct
	MultimapCount: U64
	*WireMultimap: WireMultimap
}
*/

// Serialize the schema to binary format.
func (d *WireSchema) Serialize(dst *bytes.Buffer) error {
	if err := internal.WriteUvarint(uint64(d.MainStruct), dst); err != nil {
		return nil
	}

	if err := internal.WriteUvarint(uint64(len(d.structs)), dst); err != nil {
		return err
	}

	for i := range d.structs {
		str := d.structs[i]
		if err := str.serialize(dst); err != nil {
			return nil
		}
	}

	if err := internal.WriteUvarint(uint64(len(d.multimaps)), dst); err != nil {
		return err
	}

	for i := range d.multimaps {
		mm := d.multimaps[i]
		if err := mm.serialize(dst); err != nil {
			return nil
		}
	}

	return nil
}

// Deserialize the schema from binary format.
func (d *WireSchema) Deserialize(src *bytes.Buffer) error {
	var err error
	v, err := binary.ReadUvarint(src)
	if err != nil {
		return err
	}
	d.MainStruct = StructIndex(v)

	count, err := binary.ReadUvarint(src)
	if err != nil {
		return err
	}

	if count > MaxStructOrMultimapCount {
		return errStructOrMultimapCountLimit
	}

	d.structs = make([]WireStruct, count)
	for i := 0; i < int(count); i++ {
		if err := d.structs[i].deserialize(src); err != nil {
			return err
		}
	}

	count, err = binary.ReadUvarint(src)
	if err != nil {
		return err
	}

	if count > MaxStructOrMultimapCount {
		return errStructOrMultimapCountLimit
	}

	d.multimaps = make([]WireMultimap, count)
	for i := 0; i < int(count); i++ {
		if err := d.multimaps[i].deserialize(src); err != nil {
			return err
		}
	}

	return nil
}

type WireStruct struct {
	//Name     string        `json:"name,omitempty"`
	Idx   StructIndex
	OneOf bool
	//HasDict bool
	DictIdx DictIndex
	IsRoot  bool
	Fields  []WireStructField
}

/*
Binary serialization format:

WireStruct {
	Flag: 8
	Name: String
	/DictName: String/
	FieldCount: U64
	*Field: WireFieldType
}
*/

type structFlag byte

const (
	structFlagIsRoot structFlag = 1 << iota
	structFlagOneOf
	structFlagHasDict
)

func (s *WireStruct) serialize(dst *bytes.Buffer) error {
	var flags structFlag
	if s.IsRoot {
		flags |= structFlagIsRoot
	}
	if s.OneOf {
		flags |= structFlagOneOf
	}
	if s.DictIdx != DictNone {
		flags |= structFlagHasDict
	}
	if err := dst.WriteByte(byte(flags)); err != nil {
		return err
	}
	if err := internal.WriteUvarint(uint64(s.Idx), dst); err != nil {
		return err
	}
	if s.DictIdx != DictNone {
		if err := internal.WriteUvarint(uint64(s.DictIdx), dst); err != nil {
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

func (s *WireStruct) deserialize(buf *bytes.Buffer) error {
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

	v, err := binary.ReadUvarint(buf)
	if err != nil {
		return err
	}
	s.Idx = StructIndex(v)

	if flags&structFlagHasDict != 0 {
		v, err := binary.ReadUvarint(buf)
		if err != nil {
			return err
		}
		s.DictIdx = DictIndex(v)
	}

	count, err := binary.ReadUvarint(buf)
	if err != nil {
		return err
	}

	if count > MaxStructFieldCount {
		return errStructFieldCountLimit
	}

	s.Fields = make([]WireStructField, count)
	for i := range s.Fields {
		if err := s.Fields[i].deserialize(buf); err != nil {
			return err
		}
	}

	return nil
}

type WireStructField struct {
	WireFieldType
	//Name     string `json:"name,omitempty"`
	Optional bool `json:"optional,omitempty"`
}

func (f *WireStructField) serialize(buf *bytes.Buffer) error {
	return f.WireFieldType.serialize(buf, f.Optional)
}

func (f *WireStructField) deserialize(buf *bytes.Buffer) error {
	var err error
	f.Optional, err = f.WireFieldType.deserialize(buf)
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

type WireFieldType struct {
	Primitive *PrimitiveFieldType
	Array     *WireFieldType
	Struct    StructIndex
	MultiMap  MultimapIndex
	DictIdx   DictIndex

	//DictName  string              `json:"dict,omitempty"`
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

WireFieldType {
	TypeDescr: 8
	/ElemType: WireFieldType/
	/StructName: String/
	/MultimapName: String/
	/DictName: String/
}
*/

func (f *WireFieldType) serialize(buf *bytes.Buffer, optional bool) error {
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
	} else if f.Struct != StructNone {
		typeDescr = TypeDescrStruct
	} else if f.MultiMap != MultimapNone {
		typeDescr = TypeDescrMultimap
	} else {
		panic("unknown field type")
	}

	if optional {
		typeDescr |= TypeDescrOptional
	}

	if f.DictIdx != DictNone {
		typeDescr |= TypeDescrHasDict
	}

	if err := buf.WriteByte(byte(typeDescr)); err != nil {
		return err
	}

	if f.Array != nil {
		return f.Array.serialize(buf, false)
	} else if f.Struct != StructNone {
		if err := internal.WriteUvarint(uint64(f.Struct), buf); err != nil {
			return err
		}
	} else if f.MultiMap != MultimapNone {
		if err := internal.WriteUvarint(uint64(f.MultiMap), buf); err != nil {
			return err
		}
	}

	if f.DictIdx != DictNone {
		if err := internal.WriteUvarint(uint64(f.DictIdx), buf); err != nil {
			return err
		}
	}

	return nil
}

func (f *WireFieldType) deserialize(buf *bytes.Buffer) (optional bool, err error) {
	td, err := buf.ReadByte()
	if err != nil {
		return false, err
	}
	typeDescr := TypeDescr(td)

	optional = typeDescr&TypeDescrOptional != 0
	typ := typeDescr & TypeDescrTypeMask
	switch typ {
	case TypeDescrArray:
		var elemType WireFieldType
		_, err = elemType.deserialize(buf)
		if err != nil {
			return false, err
		}
		f.Array = &elemType

	case TypeDescrStruct:
		v, err := binary.ReadUvarint(buf)
		if err != nil {
			return false, err
		}
		f.Struct = StructIndex(v)

	case TypeDescrMultimap:
		v, err := binary.ReadUvarint(buf)
		if err != nil {
			return false, err
		}
		f.MultiMap = MultimapIndex(v)

	default:
		if byte(typ) <= byte(PrimitiveTypeBytes) {
			p := PrimitiveFieldType(typ)
			f.Primitive = &p
		} else {
			return false, errors.New("unknown type")
		}
	}

	if typeDescr&TypeDescrHasDict != 0 {
		v, err := binary.ReadUvarint(buf)
		if err != nil {
			return false, err
		}
		f.DictIdx = DictIndex(v)
	}

	return optional, nil
}

type MultimapField struct {
	Type WireFieldType
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

type WireMultimap struct {
	Idx   MultimapIndex
	Key   MultimapField
	Value MultimapField
}

func (m *WireMultimap) minify() {
	//m.Name = ""
	m.Key.minify()
	m.Value.minify()
}

/*
Binary serialization format:

WireMultimap {
	Name: String
	Key: WireFieldType
	Value: WireFieldType
}
*/

func (m *WireMultimap) serialize(buf *bytes.Buffer) error {
	if err := internal.WriteUvarint(uint64(m.Idx), buf); err != nil {
		return err
	}
	if err := m.Key.serialize(buf); err != nil {
		return err
	}
	return m.Value.serialize(buf)
}

func (m *WireMultimap) deserialize(buf *bytes.Buffer) error {
	var err error
	v, err := binary.ReadUvarint(buf)
	if err != nil {
		return err
	}
	m.Idx = MultimapIndex(v)

	if err := m.Key.deserialize(buf); err != nil {
		return err
	}
	return m.Value.deserialize(buf)
}
