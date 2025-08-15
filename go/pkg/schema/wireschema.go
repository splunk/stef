package schema

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"

	"github.com/splunk/stef/go/pkg/internal"
)

// WireSchema caries only the parts of the schema, which are necessary to be
// communicated between readers and writers that work with evolving versions
// of the same schema.
//
// WireSchema allows readers and writers to perform compatibility checks
// of their schema version with the schema version that a peer they communicate
// with uses.
//
// The only valid way to evolve a STEF schema is by adding new fields at the end
// of the existing structs. This means that in order to correctly read/write an
// evolved schema the only necessary information is the number of the fields in
// in each struct. This is precisely the information that is recorded in WireSchema.
//
// The full schema information can be recorded in a schema.Schema, however that
// full information is not necessary for wire compatibility checks. Instead, we
// use the much simpler and more compact WireSchema to serve that role.
type WireSchema struct {
	// Old format:
	// Number of fields in each struct (by struct name)
	StructFieldCount map[string]uint

	// New format:
	structCounts []uint

	// Iterator:
	structCountIter int
}

// NewWireSchema creates a new WireSchema from a schema for the given root.
// The schema will be pruned to only include the structs that are reachable
// from the root struct.
func NewWireSchema(schema *Schema, root string) WireSchema {
	w := WireSchema{
		StructFieldCount: make(map[string]uint),
	}
	for k, v := range schema.Structs {
		w.StructFieldCount[k] = uint(len(v.Fields))
	}

	struc := schema.Structs[root]
	stack := recurseStack{asMap: map[string]bool{}}

	rootType := &FieldType{
		Struct:    struc.Name,
		StructDef: struc,
		DictName:  struc.DictName,
	}
	structCounts := structCountTree{}

	schemaToStructCount(rootType, &structCounts, &stack)
	w.setStructCounts(&structCounts)

	return w
}

func (w *WireSchema) FieldCount(structName string) (uint, bool) {
	count, ok := w.StructFieldCount[structName]

	if w.structCountIter >= len(w.structCounts) {
		// TODO: change return type to (uint, error) to return proper error message.
		return 0, false
	}

	count2 := w.structCounts[w.structCountIter]
	if count != count2 {
		panic(
			fmt.Sprintf(
				"FieldCount struct name mismatch: expected %s, got %d vs %d", structName, count, count2,
			),
		)
	}
	w.structCountIter++

	return count, ok
}

const (
	MaxStructOrMultimapCount = 1024
)

var (
	errStructCountLimit = errors.New("struct count limit exceeded")
)

/*
Serialization format:

WireSchema {
	StructCount:   U64
	*Struct:       WireStruct
}
WireStruct {
	StructName:    String
	FieldCount:    U64
}
String {
	LengthInBytes: U64
	*Bytes:        8
}
*/

// Serialize the schema to binary format.
func (w *WireSchema) Serialize(dst *bytes.Buffer) error {
	if err := internal.WriteUvarint(uint64(len(w.StructFieldCount)), dst); err != nil {
		return err
	}

	// Sort for deterministic serialization.
	var structs []string
	for name := range w.StructFieldCount {
		structs = append(structs, name)
	}
	sort.Strings(structs)

	// Old format
	for _, structName := range structs {
		fieldCount := w.StructFieldCount[structName]

		if err := internal.WriteString(structName, dst); err != nil {
			return err
		}
		if err := internal.WriteUvarint(uint64(fieldCount), dst); err != nil {
			return err
		}
	}

	// New format
	if err := internal.WriteUvarint(uint64(len(w.structCounts)), dst); err != nil {
		return err
	}
	for _, count := range w.structCounts {
		if err := internal.WriteUvarint(uint64(count), dst); err != nil {
			return err
		}
	}

	return nil
}

// Deserialize the schema from binary format.
func (w *WireSchema) Deserialize(src *bytes.Buffer) error {
	// Old format
	count, err := binary.ReadUvarint(src)
	if err != nil {
		return err
	}

	if count > MaxStructOrMultimapCount {
		return errStructCountLimit
	}

	w.StructFieldCount = make(map[string]uint, count)
	for i := 0; i < int(count); i++ {
		structName, err := internal.ReadString(src)
		if err != nil {
			return err
		}
		fieldCount, err := binary.ReadUvarint(src)
		if err != nil {
			return err
		}

		w.StructFieldCount[structName] = uint(fieldCount)
	}

	// New format
	count, err = binary.ReadUvarint(src)
	if err != nil {
		return err
	}

	if count > MaxStructOrMultimapCount {
		return errStructCountLimit
	}

	w.structCounts = make([]uint, int(count))

	for i := 0; i < int(count); i++ {
		fieldCount, err := binary.ReadUvarint(src)
		if err != nil {
			return err
		}

		w.structCounts[i] = uint(fieldCount)
	}

	return nil
}

// Compatible checks backward compatibility of this schema with oldSchema.
// If the schemas are incompatible returns CompatibilityIncompatible and an error.
func (w *WireSchema) Compatible(oldSchema *WireSchema) (Compatibility, error) {
	exactCompat := true
	for structName, fieldCount := range oldSchema.StructFieldCount {
		newCount, exists := w.StructFieldCount[structName]
		if !exists {
			return CompatibilityIncompatible,
				fmt.Errorf("struct %s does not exist in new schema", structName)
		}
		if newCount < fieldCount {
			return CompatibilityIncompatible,
				fmt.Errorf(
					"struct %s has fewer fields in new schema (%d vs %d)", structName,
					newCount, fieldCount,
				)
		} else if newCount > fieldCount {
			exactCompat = false
		}
	}

	if exactCompat {
		return CompatibilityExact, nil
	}

	return CompatibilitySuperset, nil
}

func (w *WireSchema) setStructCounts(s *structCountTree) {
	w.structCounts = append(w.structCounts, s.fieldCount)
	for i := range s.structFields {
		w.setStructCounts(&s.structFields[i])
	}
}
