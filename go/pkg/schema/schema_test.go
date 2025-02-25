package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"testing"

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

	minifiedJson, err := json.Marshal(prunedSchema)
	require.NoError(t, err)

	compressedJson := compressZstd(minifiedJson)

	fmt.Printf("JSON: %5d, zstd: %4d\n", len(minifiedJson), len(compressedJson))

	wireSchema := prunedSchema.ToWire()
	var wireBytes bytes.Buffer
	err = wireSchema.Serialize(&wireBytes)
	require.NoError(t, err)

	compressedBin := compressZstd(wireBytes.Bytes())
	fmt.Printf("WIRE: %5d, zstd: %4d\n", wireBytes.Len(), len(compressedBin))

	var readSchema WireSchema
	err = readSchema.Deserialize(&wireBytes)
	require.NoError(t, err)

	diff := cmp.Diff(wireSchema, readSchema)
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

		wireSchema := prunedSchema.ToWire()
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
	p := PrimitiveTypeString
	schemas := []*Schema{
		{
			PackageName: "pkg",
			Structs: map[string]*Struct{
				"Root": {Name: "Root"},
			},
		},
		{
			PackageName: "pkg",
			Structs: map[string]*Struct{
				"Root": {
					Name: "Root",
					Fields: []StructField{
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
		wireSchema := schema.ToWire()
		compat, err := wireSchema.Compatible(&wireSchema)
		require.NoError(t, err)
		assert.EqualValues(t, CompatibilityExact, compat)
	}
}

func TestSchemaSuperset(t *testing.T) {
	primitiveTypeInt64 := PrimitiveTypeInt64
	primitiveTypeString := PrimitiveTypeString

	tests := []struct {
		old *Schema
		new *Schema
	}{
		{
			old: &Schema{
				PackageName: "abc",
				Structs: map[string]*Struct{
					"Root": {
						Name: "Root",
						Fields: []StructField{
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
				PackageName: "def",
				Structs: map[string]*Struct{
					"Root": {
						Name: "Root",
						Fields: []StructField{
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
				PackageName: "abc",
				Structs: map[string]*Struct{
					"Root": {
						Name: "Root",
						Fields: []StructField{
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
						Fields: []StructField{
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
						Fields: []StructField{
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
				PackageName: "def",
				Structs: map[string]*Struct{
					"Root": {
						Name: "Root",
						Fields: []StructField{
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
						Fields: []StructField{
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
						Fields: []StructField{
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
						Fields: []StructField{
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
		oldSchema := test.old.ToWire()
		newSchema := test.new.ToWire()

		compat, err := newSchema.Compatible(&oldSchema)
		require.NoError(t, err)
		assert.EqualValues(t, CompatibilitySuperset, compat)
	}
}

func TestSchemaIncompatible(t *testing.T) {
	primitiveTypeInt64 := PrimitiveTypeInt64

	tests := []struct {
		old *Schema
		new *Schema
		err string
	}{
		{
			old: &Schema{
				PackageName: "abc",
				Structs: map[string]*Struct{
					"Root": {
						Name: "Root",
						Fields: []StructField{
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
				PackageName: "def",
				Structs: map[string]*Struct{
					"Root": {
						Name: "Root",
						Fields: []StructField{
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
			err: "struct Root has fewer fields in new schema (1 vs 2)",
		},
	}

	for _, test := range tests {
		oldSchema := test.old.ToWire()
		newSchema := test.new.ToWire()

		compat, err := newSchema.Compatible(&oldSchema)
		require.Error(t, err)
		assert.EqualValues(t, test.err, err.Error())
		assert.EqualValues(t, CompatibilityIncompatible, compat)
	}
}

func expandSchema(t *testing.T, r *rand.Rand, orig *Schema) (cpy *Schema) {
	cpy, err := orig.PrunedForRoot("Metrics")
	require.NoError(t, err)
	for {
		for _, str := range cpy.Structs {
			if expandStruct(t, r, cpy, str) {
				return cpy
			}
		}
	}
}

func expandStruct(t *testing.T, r *rand.Rand, schema *Schema, str *Struct) bool {
	if r.Intn(10) == 0 {
		field := StructField{
			FieldType: FieldType{},
			Name:      fmt.Sprintf("Field#%d", len(str.Fields)+1),
		}

		p := PrimitiveTypeString
		switch r.Intn(4) {
		case 0:
			field.FieldType.Primitive = &p
			if r.Intn(10) == 0 {
				field.DictName = "Dict" + field.Name
			}

		case 1:
			f := FieldType{Primitive: &p}
			field.FieldType.Array = &f
		case 2:
			multimapIdx := r.Intn(len(schema.Multimaps))
			i := 0
			for multimapName := range schema.Multimaps {
				if i == multimapIdx {
					field.FieldType.MultiMap = multimapName
					break
				}
				i++
			}
		case 3:
			if r.Intn(2) == 0 {
				// Add new struct
				struc := Struct{
					Name:   fmt.Sprintf("Struct#%d", len(schema.Structs)),
					Fields: []StructField{},
				}
				schema.Structs[struc.Name] = &struc
				field.FieldType.Struct = struc.Name
			} else {
				structIdx := r.Intn(len(schema.Structs))
				i := 0
				for structName := range schema.Structs {
					if i == structIdx {
						field.FieldType.Struct = structName
						break
					}
					i++
				}
			}
		}

		str.Fields = append(str.Fields, field)
		return true
	}

	for _, field := range str.Fields {
		if field.Struct != "" {
			if r.Intn(10) == 0 {
				childStruct := schema.Structs[field.Struct]
				changed := expandStruct(t, r, schema, childStruct)
				if changed {
					return true
				}
			}
		}
	}

	return false
}

func shrinkSchema(t *testing.T, r *rand.Rand, orig *Schema) (cpy *Schema) {
	cpy, err := orig.PrunedForRoot("Metrics")
	require.NoError(t, err)
	for {
		for _, str := range cpy.Structs {
			if shrinkStruct(t, r, cpy, str) {
				return cpy
			}
		}
	}
}

func shrinkStruct(t *testing.T, r *rand.Rand, schema *Schema, str *Struct) bool {
	if r.Intn(10) == 0 && len(str.Fields) > 0 {
		str.Fields = str.Fields[0 : len(str.Fields)-1]
		return true
	}

	for _, field := range str.Fields {
		if field.Struct != "" {
			if r.Intn(3) == 0 {
				childStruct := schema.Structs[field.Struct]
				changed := shrinkStruct(t, r, schema, childStruct)
				if changed {
					return true
				}
			}
		}
	}

	return false
}

func TestSchemaExpand(t *testing.T) {
	schemaJson, err := os.ReadFile("testdata/oteltef.wire.json")
	require.NoError(t, err)

	orig := &Schema{}
	err = json.Unmarshal(schemaJson, &orig)
	require.NoError(t, err)
	orig, err = orig.PrunedForRoot("Metrics")
	require.NoError(t, err)

	r := rand.New(rand.NewSource(42))

	// Expand one field at a time and check compatibility.
	for i := 0; i < 200; i++ {
		expanded := expandSchema(t, r, orig)
		expandedWire := expanded.ToWire()
		require.NoError(t, err)

		// Exact compatible with itself
		compat, err := expandedWire.Compatible(&expandedWire)
		require.NoError(t, err, i)
		assert.EqualValues(t, CompatibilityExact, compat, i)

		// Expanding is compatible / superset
		origWire := orig.ToWire()
		require.NoError(t, err, i)
		compat, err = expandedWire.Compatible(&origWire)
		require.NoError(t, err, i)
		assert.EqualValues(t, CompatibilitySuperset, compat, i)

		// Opposite direction is incompatible
		compat, err = origWire.Compatible(&expandedWire)
		require.Error(t, err, i)
		assert.EqualValues(t, CompatibilityIncompatible, compat, i)

		// Also check that serialization works correctly.

		// Serialize
		var buf bytes.Buffer
		err = expandedWire.Serialize(&buf)
		require.NoError(t, err)

		// Deserialize
		var cpy WireSchema
		err = cpy.Deserialize(&buf)
		require.NoError(t, err)

		// Compare deserialized schema
		require.EqualValues(t, expandedWire, cpy)

		orig = expanded
	}
}

func TestSchemaShrink(t *testing.T) {
	schemaJson, err := os.ReadFile("testdata/oteltef.wire.json")
	require.NoError(t, err)

	orig := &Schema{}
	err = json.Unmarshal(schemaJson, &orig)
	require.NoError(t, err)

	r := rand.New(rand.NewSource(42))

	// Expand the schema, make it much bigger, so there is room for shrinking.
	for i := 0; i < 200; i++ {
		orig = expandSchema(t, r, orig)
	}

	// Now shrink one field at a time and check compatibility.
	for i := 0; i < 100; i++ {
		shrinked := shrinkSchema(t, r, orig)
		shrinkedWire := shrinked.ToWire()
		require.NoError(t, err)

		// Shrinking is incompatible
		origWire := orig.ToWire()
		compat, err := shrinkedWire.Compatible(&origWire)
		require.Error(t, err, i)
		assert.EqualValues(t, CompatibilityIncompatible, compat, i)

		// Opposite direction is compatible/superset
		compat, err = origWire.Compatible(&shrinkedWire)
		require.NoError(t, err, i)
		assert.EqualValues(t, CompatibilitySuperset, compat, i)

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
}
