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

	prunedSchema.Minify()
	minifiedJson, err := json.Marshal(prunedSchema)
	require.NoError(t, err)

	compressedJson := compressZstd(minifiedJson)

	fmt.Printf("JSON: %5d, zstd: %4d\n", len(minifiedJson), len(compressedJson))

	var wireBytes bytes.Buffer
	err = prunedSchema.Serialize(&wireBytes)
	require.NoError(t, err)

	compressedBin := compressZstd(wireBytes.Bytes())
	fmt.Printf("BIN: %5d, zstd: %4d\n", wireBytes.Len(), len(compressedBin))

	var readSchema Schema
	err = readSchema.Deserialize(&wireBytes)
	require.NoError(t, err)

	diff := cmp.Diff(prunedSchema, &readSchema)
	if diff != "" {
		assert.Fail(t, diff)
	}

	assert.True(t, reflect.DeepEqual(prunedSchema, &readSchema))
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

		prunedSchema.Minify()

		var wireBytes bytes.Buffer
		err = prunedSchema.Serialize(&wireBytes)
		require.NoError(f, err)

		f.Add(wireBytes.Bytes())
	}

	f.Fuzz(
		func(t *testing.T, data []byte) {
			var readSchema Schema
			_ = readSchema.Deserialize(bytes.NewBuffer(data))
		},
	)
}

func TestSchemaSelfCompatible(t *testing.T) {
	schemas := []*Schema{
		{
			PackageName: "pkg",
			Structs: map[string]*Struct{
				"Root": {Name: "Root"},
			},
			MainStruct: "Root",
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
				"Multi": {Name: "Multi"},
			},
			MainStruct: "Root",
		},
	}

	for _, schema := range schemas {
		compat, err := schema.Compatible(schema)
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
				Multimaps:  nil,
				MainStruct: "Root",
			},
			new: &Schema{
				PackageName: "def",
				Structs: map[string]*Struct{
					"Root2": {
						Name: "Root2",
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
				Multimaps:  nil,
				MainStruct: "Root2",
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
									Struct: "Aold",
								},
								Name: "F2",
							},
						},
					},
					"Aold": {
						Name: "Aold",
						Fields: []StructField{
							{
								FieldType: FieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
							{
								FieldType: FieldType{
									Struct: "Bold",
								},
								Name:     "F2",
								Optional: true,
							},
							{
								FieldType: FieldType{
									MultiMap: "Mold",
								},
								Name: "F3",
							},
						},
					},
					"Bold": {
						Name: "Bold",
						Fields: []StructField{
							{
								FieldType: FieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
							{
								FieldType: FieldType{
									Struct: "Aold",
								},
								Name: "F2",
							},
						},
					},
				},
				Multimaps: map[string]*Multimap{
					"Mold": {
						Name:  "Mold",
						Key:   MultimapField{Type: FieldType{Primitive: &primitiveTypeInt64}},
						Value: MultimapField{Type: FieldType{Primitive: &primitiveTypeString}},
					},
				},
				MainStruct: "Root",
			},
			new: &Schema{
				PackageName: "def",
				Structs: map[string]*Struct{
					"Root2": {
						Name: "Root2",
						Fields: []StructField{
							{
								FieldType: FieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
							{
								FieldType: FieldType{
									Struct: "Anew",
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
					"Anew": {
						// This corresponds to "Aold" in old schema
						Name: "Anew",
						Fields: []StructField{
							{
								FieldType: FieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
							{
								FieldType: FieldType{
									Struct: "Bnew",
								},
								Name:     "F2",
								Optional: true,
							},
							{
								FieldType: FieldType{
									MultiMap: "Mnew",
								},
								Name: "F3",
							},
						},
					},
					"Bnew": {
						// This corresponds to "Bold" in old schema.
						Name: "Bnew",
						Fields: []StructField{
							{
								FieldType: FieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
							{
								FieldType: FieldType{
									Struct: "Anew",
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
					"Mnew": {
						Name:  "Mnew",
						Key:   MultimapField{Type: FieldType{Primitive: &primitiveTypeInt64}},
						Value: MultimapField{Type: FieldType{Primitive: &primitiveTypeString}},
					},
				},
				MainStruct: "Root2",
			},
		},
	}

	for _, test := range tests {
		compat, err := test.new.Compatible(test.old)
		require.NoError(t, err)
		assert.EqualValues(t, CompatibilitySuperset, compat)
	}
}

func TestSchemaIncompatible(t *testing.T) {
	primitiveTypeInt64 := PrimitiveTypeInt64
	primitiveTypeString := PrimitiveTypeString

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
						},
					},
				},
				Multimaps:  nil,
				MainStruct: "Root",
			},
			new: &Schema{
				PackageName: "def",
				Structs: map[string]*Struct{
					"Root": {
						Name: "Root",
						Fields: []StructField{
							{
								FieldType: FieldType{
									Primitive: &primitiveTypeString,
								},
								Name: "F1",
							},
						},
					},
				},
				MainStruct: "Root",
			},
			err: "field 0 in new struct Root has a different type than in the old struct Root",
		},
	}

	for _, test := range tests {
		compat, err := test.new.Compatible(test.old)
		require.Error(t, err)
		assert.EqualValues(t, test.err, err.Error())
		assert.EqualValues(t, CompatibilityIncompatible, compat)
	}
}

func expandSchema(t *testing.T, r *rand.Rand, orig *Schema) (cpy *Schema) {
	cpy, err := orig.PrunedForRoot(orig.MainStruct)
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
			//Name:      fmt.Sprintf("Field#%d", len(str.Fields)+1),
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
	cpy, err := orig.PrunedForRoot(orig.MainStruct)
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

	orig.Minify()

	r := rand.New(rand.NewSource(42))

	// Expand one field at a time and check compatibility.
	for i := 0; i < 200; i++ {
		expanded := expandSchema(t, r, orig)

		// Exact compatible with itself
		compat, err := expanded.Compatible(expanded)
		require.NoError(t, err, i)
		assert.EqualValues(t, CompatibilityExact, compat, i)

		// Expanding is compatible / superset
		compat, err = expanded.Compatible(orig)
		require.NoError(t, err, i)
		assert.EqualValues(t, CompatibilitySuperset, compat, i)

		// Opposite direction is incompatible
		compat, err = orig.Compatible(expanded)
		require.Error(t, err, i)
		assert.EqualValues(t, CompatibilityIncompatible, compat, i)

		// Also check that serialization works correctly.

		// Serialize
		var buf bytes.Buffer
		err = expanded.Serialize(&buf)
		require.NoError(t, err)

		// Deserialize
		var cpy Schema
		err = cpy.Deserialize(&buf)
		require.NoError(t, err)

		// Compare deserialized schema
		require.EqualValues(t, expanded, &cpy)

		orig = expanded
	}
}

func TestSchemaShrink(t *testing.T) {
	schemaJson, err := os.ReadFile("testdata/oteltef.wire.json")
	require.NoError(t, err)

	orig := &Schema{}
	err = json.Unmarshal(schemaJson, &orig)
	require.NoError(t, err)

	orig.Minify()

	r := rand.New(rand.NewSource(42))

	// Expand the schema, make it much bigger, so there is room for shrinking.
	for i := 0; i < 200; i++ {
		orig = expandSchema(t, r, orig)
	}

	// Now shrink one field at a time and check compatibility.
	for i := 0; i < 100; i++ {
		shrinked := shrinkSchema(t, r, orig)

		// Shrinking is incompatible
		compat, err := shrinked.Compatible(orig)
		require.Error(t, err, i)
		assert.EqualValues(t, CompatibilityIncompatible, compat, i)

		// Opposite direction is compatible/superset
		compat, err = orig.Compatible(shrinked)
		require.NoError(t, err, i)
		assert.EqualValues(t, CompatibilitySuperset, compat, i)

		// Also check that serialization works correctly.

		// Serialize
		var buf bytes.Buffer
		err = shrinked.Serialize(&buf)
		require.NoError(t, err)

		// Deserialize
		var cpy Schema
		err = cpy.Deserialize(&buf)
		require.NoError(t, err)

		// Compare deserialized schema
		require.EqualValues(t, shrinked, &cpy)

		orig = shrinked
	}
}
