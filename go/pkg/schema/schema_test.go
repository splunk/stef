package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var zstdEncoder, _ = zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))

func compressZstd(input []byte) []byte {
	return zstdEncoder.EncodeAll(input, make([]byte, 0, len(input)))
}

func TestSerializeSchema(t *testing.T) {
	wireJson, err := os.ReadFile("testdata/example.json")
	require.NoError(t, err)

	var schema Schema
	err = json.Unmarshal(wireJson, &schema)
	require.NoError(t, err)

	prunedSchema, err := schema.PrunedForRoot("Metrics")
	require.NoError(t, err)

	wireSchema := NewWireSchema(prunedSchema, "Metrics")
	var wireBytes bytes.Buffer
	err = wireSchema.Serialize(&wireBytes)
	require.NoError(t, err)

	compressedBin := compressZstd(wireBytes.Bytes())
	fmt.Printf("WIRE: %5d, zstd: %4d\n", wireBytes.Len(), len(compressedBin))

	var readSchema WireSchema
	err = readSchema.Deserialize(&wireBytes)
	require.NoError(t, err)

	diff := cmp.Diff(wireSchema, readSchema, cmp.AllowUnexported(WireSchema{}))
	if diff != "" {
		assert.Fail(t, diff)
	}

	assert.True(t, reflect.DeepEqual(wireSchema, readSchema))
}

func FuzzDeserialize(f *testing.F) {
	wireJson, err := os.ReadFile("testdata/example.json")
	require.NoError(f, err)

	var schema Schema
	err = json.Unmarshal(wireJson, &schema)
	require.NoError(f, err)

	roots := []string{"Metrics", "Spans"}

	for _, root := range roots {
		prunedSchema, err := schema.PrunedForRoot(root)
		require.NoError(f, err)

		wireSchema := NewWireSchema(prunedSchema, root)
		var wireBytes bytes.Buffer
		err = wireSchema.Serialize(&wireBytes)
		require.NoError(f, err)

		f.Add(wireBytes.Bytes())
	}

	f.Fuzz(
		func(t *testing.T, data []byte) {
			var readSchema WireSchema
			_ = readSchema.Deserialize(bytes.NewBuffer(data))
		},
	)
}

func TestSchemaSelfCompatible(t *testing.T) {
	p := PrimitiveType{Type: PrimitiveTypeString}
	schemas := []*Schema{
		{
			PackageName: []string{"pkg"},
			Structs: map[string]*Struct{
				"Root": {Name: "Root"},
			},
		},
		{
			PackageName: []string{"pkg"},
			Structs: map[string]*Struct{
				"Root": {
					Name: "Root",
					Fields: []*StructField{
						{
							FieldType: FieldType{MultiMap: "Multi"},
							Name:      "F1",
						},
					},
				},
			},
			Multimaps: map[string]*Multimap{
				"Multi": {
					Name:  "Multi",
					Key:   MultimapField{Type: FieldType{Primitive: &p}},
					Value: MultimapField{Type: FieldType{Primitive: &p}},
				},
			},
		},
	}

	for _, schema := range schemas {
		err := schema.ResolveRefs()
		require.NoError(t, err)

		wireSchema := NewWireSchema(schema, "Root")
		compat, err := wireSchema.Compatible(&wireSchema)
		require.NoError(t, err)
		assert.EqualValues(t, CompatibilityExact, compat)
	}
}

func TestSchemaSuperset(t *testing.T) {
	primitiveTypeInt64 := PrimitiveType{Type: PrimitiveTypeInt64}
	primitiveTypeString := PrimitiveType{Type: PrimitiveTypeString}

	tests := []struct {
		old *Schema
		new *Schema
	}{
		{
			old: &Schema{
				PackageName: []string{"abc"},
				Structs: map[string]*Struct{
					"Root": {
						Name: "Root",
						Fields: []*StructField{
							{
								FieldType: FieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
						},
					},
				},
				Multimaps: nil,
			},
			new: &Schema{
				PackageName: []string{"def"},
				Structs: map[string]*Struct{
					"Root": {
						Name: "Root",
						Fields: []*StructField{
							{
								FieldType: FieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
							{
								FieldType: FieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F2",
							},
						},
					},
				},
				Multimaps: nil,
			},
		},
		{
			old: &Schema{
				PackageName: []string{"abc"},
				Structs: map[string]*Struct{
					"Root": {
						Name: "Root",
						Fields: []*StructField{
							{
								FieldType: FieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
							{
								FieldType: FieldType{
									Struct: "A",
								},
								Name: "F2",
							},
						},
					},
					"A": {
						Name: "A",
						Fields: []*StructField{
							{
								FieldType: FieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
							{
								FieldType: FieldType{
									Struct: "B",
								},
								Name:     "F2",
								Optional: true,
							},
							{
								FieldType: FieldType{
									MultiMap: "M",
								},
								Name: "F3",
							},
						},
					},
					"B": {
						Name: "B",
						Fields: []*StructField{
							{
								FieldType: FieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
							{
								FieldType: FieldType{
									Struct: "A",
								},
								Name: "F2",
							},
						},
					},
				},
				Multimaps: map[string]*Multimap{
					"M": {
						Name:  "M",
						Key:   MultimapField{Type: FieldType{Primitive: &primitiveTypeInt64}},
						Value: MultimapField{Type: FieldType{Primitive: &primitiveTypeString}},
					},
				},
			},
			new: &Schema{
				PackageName: []string{"def"},
				Structs: map[string]*Struct{
					"Root": {
						Name: "Root",
						Fields: []*StructField{
							{
								FieldType: FieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
							{
								FieldType: FieldType{
									Struct: "A",
								},
								Name: "F2",
							},
							{
								FieldType: FieldType{
									Struct: "D",
								},
								Name: "F3",
							},
						},
					},
					"A": {
						Name: "A",
						Fields: []*StructField{
							{
								FieldType: FieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
							{
								FieldType: FieldType{
									Struct: "B",
								},
								Name:     "F2",
								Optional: true,
							},
							{
								FieldType: FieldType{
									MultiMap: "M",
								},
								Name: "F3",
							},
						},
					},
					"B": {
						Name: "B",
						Fields: []*StructField{
							{
								FieldType: FieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
							{
								FieldType: FieldType{
									Struct: "A",
								},
								Name: "F2",
							},
							{
								FieldType: FieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F3",
							},
						},
					},
					"D": {
						Name:     "D",
						OneOf:    true,
						DictName: "",
						Fields: []*StructField{
							{
								FieldType: FieldType{Primitive: &primitiveTypeInt64},
								Name:      "F1",
							},
						},
					},
				},
				Multimaps: map[string]*Multimap{
					"M": {
						Name:  "M",
						Key:   MultimapField{Type: FieldType{Primitive: &primitiveTypeInt64}},
						Value: MultimapField{Type: FieldType{Primitive: &primitiveTypeString}},
					},
				},
			},
		},
	}

	for _, test := range tests {
		err := test.old.ResolveRefs()
		require.NoError(t, err)
		err = test.new.ResolveRefs()
		require.NoError(t, err)

		oldSchema := NewWireSchema(test.old, "Root")
		newSchema := NewWireSchema(test.new, "Root")

		compat, err := newSchema.Compatible(&oldSchema)
		require.NoError(t, err)
		assert.EqualValues(t, CompatibilitySuperset, compat)
	}
}

func TestSchemaIncompatible(t *testing.T) {
	primitiveTypeInt64 := PrimitiveType{Type: PrimitiveTypeInt64}

	tests := []struct {
		old *Schema
		new *Schema
		err string
	}{
		{
			old: &Schema{
				PackageName: []string{"abc"},
				Structs: map[string]*Struct{
					"Root": {
						Name: "Root",
						Fields: []*StructField{
							{
								FieldType: FieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
							{
								FieldType: FieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F2",
							},
						},
					},
				},
				Multimaps: nil,
			},
			new: &Schema{
				PackageName: []string{"def"},
				Structs: map[string]*Struct{
					"Root": {
						Name: "Root",
						Fields: []*StructField{
							{
								FieldType: FieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
						},
					},
				},
			},
			err: "new schema has fewers fields than old schema (1 vs 2)",
		},
	}

	for _, test := range tests {
		err := test.old.ResolveRefs()
		require.NoError(t, err)
		err = test.new.ResolveRefs()
		require.NoError(t, err)

		oldSchema := NewWireSchema(test.old, "Root")
		newSchema := NewWireSchema(test.new, "Root")

		compat, err := newSchema.Compatible(&oldSchema)
		require.Error(t, err)
		assert.EqualValues(t, test.err, err.Error())
		assert.EqualValues(t, CompatibilityIncompatible, compat)
	}
}

func expandSchema(t *testing.T, r *rand.Rand, orig *Schema) (cpy *Schema) {
	cpy, err := orig.PrunedForRoot("Metrics")
	require.NoError(t, err)

	var structNames []string
	for name := range cpy.Structs {
		structNames = append(structNames, name)
	}
	sort.Strings(structNames)

	var multimapNames []string
	for name := range cpy.Multimaps {
		multimapNames = append(multimapNames, name)
	}
	sort.Strings(multimapNames)

	for {
		structName := structNames[r.IntN(len(structNames))]
		str := cpy.Structs[structName]
		if expandStruct(t, r, cpy, str, structNames, multimapNames) {
			err = cpy.ResolveRefs()
			require.NoError(t, err)
			return cpy
		}
	}
}

func expandStruct(t *testing.T, r *rand.Rand, schema *Schema, str *Struct, structNames, multimapNames []string) bool {
	if r.IntN(10) == 0 {
		field := StructField{
			FieldType: FieldType{},
			Name:      fmt.Sprintf("Field#%d", len(str.Fields)+1),
		}

		p := PrimitiveType{Type: PrimitiveTypeString}
		switch r.IntN(4) {
		case 0:
			field.FieldType.Primitive = &p
			if r.IntN(10) == 0 {
				field.DictName = "Dict" + field.Name
			}

		case 1:
			f := FieldType{Primitive: &p}
			field.FieldType.Array = &ArrayType{ElemType: f}
		case 2:
			field.FieldType.MultiMap = multimapNames[r.IntN(len(multimapNames))]
		case 3:
			if r.IntN(2) == 0 {
				// Add new struct
				struc := Struct{
					Name:   fmt.Sprintf("Struct#%d", len(schema.Structs)),
					Fields: []*StructField{},
				}
				schema.Structs[struc.Name] = &struc
				field.FieldType.Struct = struc.Name
			} else {
				field.FieldType.Struct = structNames[r.IntN(len(structNames))]
			}
		}

		str.Fields = append(str.Fields, &field)
		return true
	}

	for _, field := range str.Fields {
		if field.Struct != "" {
			if r.IntN(10) == 0 {
				childStruct := schema.Structs[field.Struct]
				changed := expandStruct(t, r, schema, childStruct, structNames, multimapNames)
				if changed {
					return true
				}
			}
		}
	}

	return false
}

func TestSchemaExpand(t *testing.T) {
	schemaJson, err := os.ReadFile("testdata/example.json")
	require.NoError(t, err)

	orig := &Schema{}
	err = json.Unmarshal(schemaJson, &orig)
	require.NoError(t, err)
	orig, err = orig.PrunedForRoot("Metrics")
	require.NoError(t, err)

	seed1 := uint64(time.Now().UnixNano())
	r := rand.New(rand.NewPCG(seed1, 0))

	succeeded := false
	defer func() {
		if !succeeded {
			fmt.Printf("Test failed with seed %v\n", seed1)
		}
	}()

	// Expand one field at a time and check compatibility.
	for i := 0; i < 200; i++ {
		expanded := expandSchema(t, r, orig)
		expandedWire := NewWireSchema(expanded, "Metrics")
		require.NoError(t, err)

		// Exact compatible with itself
		compat, err := expandedWire.Compatible(&expandedWire)
		require.NoError(t, err)
		require.EqualValues(t, CompatibilityExact, compat)

		// Expanding is compatible / superset
		origWire := NewWireSchema(orig, "Metrics")
		require.NoError(t, err)
		compat, err = expandedWire.Compatible(&origWire)
		require.NoError(t, err)
		require.EqualValues(t, CompatibilitySuperset, compat)

		// Opposite direction is incompatible
		compat, err = origWire.Compatible(&expandedWire)
		require.Error(t, err)
		require.EqualValues(t, CompatibilityIncompatible, compat)

		// Also check that serialization works correctly.

		// Serialize
		var buf bytes.Buffer
		err = expandedWire.Serialize(&buf)
		require.NoError(t, err)

		// Deserialize
		var cpy WireSchema
		err = cpy.Deserialize(&buf)
		require.NoError(t, err, compat)

		// Compare deserialized schema
		require.EqualValues(t, expandedWire, cpy)

		orig = expanded
	}

	succeeded = true
}

func TestSchemaShrink(t *testing.T) {
	schemaJson, err := os.ReadFile("testdata/example.json")
	require.NoError(t, err)

	orig := &Schema{}
	err = json.Unmarshal(schemaJson, &orig)
	require.NoError(t, err)

	seed1 := uint64(time.Now().UnixNano())
	r := rand.New(rand.NewPCG(seed1, 0))

	succeeded := false
	defer func() {
		if !succeeded {
			fmt.Printf("Test failed with seed %v\n", seed1)
		}
	}()

	// Expand the schema, make it much bigger, so there is room for shrinking.
	for i := 0; i < 200; i++ {
		orig = expandSchema(t, r, orig)
	}

	// Now shrink one field at a time and check compatibility.
	for i := 0; i < 100; i++ {
		// Make a copy
		shrinked, err := orig.PrunedForRoot("Metrics")
		require.NoError(t, err)

		// Srhink it
		ShrinkRandomly(r, shrinked)
		shrinkedWire := NewWireSchema(shrinked, "Metrics")
		require.NoError(t, err)

		// Shrinking is incompatible
		origWire := NewWireSchema(orig, "Metrics")
		compat, err := shrinkedWire.Compatible(&origWire)
		require.Error(t, err)
		require.EqualValues(t, CompatibilityIncompatible, compat)

		// Opposite direction is compatible/superset
		compat, err = origWire.Compatible(&shrinkedWire)
		require.NoError(t, err, "%d %d", i, seed1)
		require.EqualValues(t, CompatibilitySuperset, compat)

		// Also check that serialization works correctly.

		// Serialize
		var buf bytes.Buffer
		err = shrinkedWire.Serialize(&buf)
		require.NoError(t, err)

		// Deserialize
		var cpy WireSchema
		err = cpy.Deserialize(&buf)
		require.NoError(t, err)

		// Compare deserialized schema
		require.EqualValues(t, shrinkedWire, cpy)

		orig = shrinked
	}

	succeeded = true
}
