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

	var schema JsonSchema
	err = json.Unmarshal(wireJson, &schema)
	require.NoError(t, err)

	prunedSchema, err := schema.PrunedForRoot("Metrics")
	require.NoError(t, err)

	prunedJson, err := json.Marshal(prunedSchema)
	require.NoError(t, err)

	compressedJson := compressZstd(prunedJson)

	fmt.Printf("JSON: %5d, zstd: %4d\n", len(prunedJson), len(compressedJson))

	wireSchema, err := prunedSchema.ToWire()
	require.NoError(t, err)

	var wireBytes bytes.Buffer
	err = wireSchema.Serialize(&wireBytes)
	require.NoError(t, err)

	compressedBin := compressZstd(wireBytes.Bytes())
	fmt.Printf("BIN: %5d, zstd: %4d\n", wireBytes.Len(), len(compressedBin))

	var readSchema WireSchema
	err = readSchema.Deserialize(&wireBytes)
	require.NoError(t, err)

	diff := cmp.Diff(wireSchema, &readSchema, cmp.AllowUnexported(WireSchema{}))
	if diff != "" {
		assert.Fail(t, diff)
	}

	assert.True(t, reflect.DeepEqual(wireSchema, &readSchema))
}

func FuzzDeserialize(f *testing.F) {
	wireJson, err := os.ReadFile("testdata/example.json")
	require.NoError(f, err)

	var schema JsonSchema
	err = json.Unmarshal(wireJson, &schema)
	require.NoError(f, err)

	roots := []string{"Metrics", "Spans"}

	for _, root := range roots {
		prunedSchema, err := schema.PrunedForRoot(root)
		require.NoError(f, err)

		JsonSchema, err := prunedSchema.ToWire()
		require.NoError(f, err)

		var wireBytes bytes.Buffer
		err = JsonSchema.Serialize(&wireBytes)
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
	schemas := []*JsonSchema{
		{
			PackageName: "pkg",
			Structs: map[string]*JsonStruct{
				"Root": {Name: "Root"},
			},
			MainStruct: "Root",
		},
		{
			PackageName: "pkg",
			Structs: map[string]*JsonStruct{
				"Root": {
					Name: "Root",
					Fields: []JsonStructField{
						{
							JsonFieldType: JsonFieldType{MultiMap: "Multi"},
							Name:          "F1",
						},
					},
				},
			},
			Multimaps: map[string]*JsonMultimap{
				"Multi": {
					Name:  "Multi",
					Key:   JsonMultimapField{Type: JsonFieldType{Primitive: &p}},
					Value: JsonMultimapField{Type: JsonFieldType{Primitive: &p}},
				},
			},
			MainStruct: "Root",
		},
	}

	for _, schema := range schemas {
		JsonSchema, err := schema.ToWire()
		require.NoError(t, err)
		compat, err := JsonSchema.Compatible(JsonSchema)
		require.NoError(t, err)
		assert.EqualValues(t, CompatibilityExact, compat)
	}
}

func TestSchemaSuperset(t *testing.T) {
	primitiveTypeInt64 := PrimitiveTypeInt64
	primitiveTypeString := PrimitiveTypeString

	tests := []struct {
		old *JsonSchema
		new *JsonSchema
	}{
		{
			old: &JsonSchema{
				PackageName: "abc",
				Structs: map[string]*JsonStruct{
					"Root": {
						Name: "Root",
						Fields: []JsonStructField{
							{
								JsonFieldType: JsonFieldType{
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
			new: &JsonSchema{
				PackageName: "def",
				Structs: map[string]*JsonStruct{
					"Root2": {
						Name: "Root2",
						Fields: []JsonStructField{
							{
								JsonFieldType: JsonFieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
							{
								JsonFieldType: JsonFieldType{
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
			old: &JsonSchema{
				PackageName: "abc",
				Structs: map[string]*JsonStruct{
					"Root": {
						Name: "Root",
						Fields: []JsonStructField{
							{
								JsonFieldType: JsonFieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
							{
								JsonFieldType: JsonFieldType{
									Struct: "Aold",
								},
								Name: "F2",
							},
						},
					},
					"Aold": {
						Name: "Aold",
						Fields: []JsonStructField{
							{
								JsonFieldType: JsonFieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
							{
								JsonFieldType: JsonFieldType{
									Struct: "Bold",
								},
								Name:     "F2",
								Optional: true,
							},
							{
								JsonFieldType: JsonFieldType{
									MultiMap: "Mold",
								},
								Name: "F3",
							},
						},
					},
					"Bold": {
						Name: "Bold",
						Fields: []JsonStructField{
							{
								JsonFieldType: JsonFieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
							{
								JsonFieldType: JsonFieldType{
									Struct: "Aold",
								},
								Name: "F2",
							},
						},
					},
				},
				Multimaps: map[string]*JsonMultimap{
					"Mold": {
						Name:  "Mold",
						Key:   JsonMultimapField{Type: JsonFieldType{Primitive: &primitiveTypeInt64}},
						Value: JsonMultimapField{Type: JsonFieldType{Primitive: &primitiveTypeString}},
					},
				},
				MainStruct: "Root",
			},
			new: &JsonSchema{
				PackageName: "def",
				Structs: map[string]*JsonStruct{
					"Root2": {
						Name: "Root2",
						Fields: []JsonStructField{
							{
								JsonFieldType: JsonFieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
							{
								JsonFieldType: JsonFieldType{
									Struct: "Anew",
								},
								Name: "F2",
							},
							{
								JsonFieldType: JsonFieldType{
									Struct: "D",
								},
								Name: "F3",
							},
						},
					},
					"Anew": {
						// This corresponds to "Aold" in old schema
						Name: "Anew",
						Fields: []JsonStructField{
							{
								JsonFieldType: JsonFieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
							{
								JsonFieldType: JsonFieldType{
									Struct: "Bnew",
								},
								Name:     "F2",
								Optional: true,
							},
							{
								JsonFieldType: JsonFieldType{
									MultiMap: "Mnew",
								},
								Name: "F3",
							},
						},
					},
					"Bnew": {
						// This corresponds to "Bold" in old schema.
						Name: "Bnew",
						Fields: []JsonStructField{
							{
								JsonFieldType: JsonFieldType{
									Primitive: &primitiveTypeInt64,
								},
								Name: "F1",
							},
							{
								JsonFieldType: JsonFieldType{
									Struct: "Anew",
								},
								Name: "F2",
							},
							{
								JsonFieldType: JsonFieldType{
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
						Fields: []JsonStructField{
							{
								JsonFieldType: JsonFieldType{Primitive: &primitiveTypeInt64},
								Name:          "F1",
							},
						},
					},
				},
				Multimaps: map[string]*JsonMultimap{
					"Mnew": {
						Name:  "Mnew",
						Key:   JsonMultimapField{Type: JsonFieldType{Primitive: &primitiveTypeInt64}},
						Value: JsonMultimapField{Type: JsonFieldType{Primitive: &primitiveTypeString}},
					},
				},
				MainStruct: "Root2",
			},
		},
	}

	for _, test := range tests {
		oldSchema, err := test.old.ToWire()
		require.NoError(t, err)
		newSchema, err := test.new.ToWire()
		require.NoError(t, err)

		compat, err := newSchema.Compatible(oldSchema)
		require.NoError(t, err)
		assert.EqualValues(t, CompatibilitySuperset, compat)
	}
}

func TestSchemaIncompatible(t *testing.T) {
	primitiveTypeInt64 := PrimitiveTypeInt64
	primitiveTypeString := PrimitiveTypeString

	tests := []struct {
		old *JsonSchema
		new *JsonSchema
		err string
	}{
		{
			old: &JsonSchema{
				PackageName: "abc",
				Structs: map[string]*JsonStruct{
					"Root": {
						Name: "Root",
						Fields: []JsonStructField{
							{
								JsonFieldType: JsonFieldType{
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
			new: &JsonSchema{
				PackageName: "def",
				Structs: map[string]*JsonStruct{
					"Root": {
						Name: "Root",
						Fields: []JsonStructField{
							{
								JsonFieldType: JsonFieldType{
									Primitive: &primitiveTypeString,
								},
								Name: "F1",
							},
						},
					},
				},
				MainStruct: "Root",
			},
			err: "field 0 in new struct 1 has a different type than in the old struct 1",
		},
	}

	for _, test := range tests {
		oldSchema, err := test.old.ToWire()
		require.NoError(t, err)
		newSchema, err := test.new.ToWire()
		require.NoError(t, err)

		compat, err := newSchema.Compatible(oldSchema)
		require.Error(t, err)
		assert.EqualValues(t, test.err, err.Error())
		assert.EqualValues(t, CompatibilityIncompatible, compat)
	}
}

func expandSchema(t *testing.T, r *rand.Rand, orig *JsonSchema) (cpy *JsonSchema) {
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

func expandStruct(t *testing.T, r *rand.Rand, schema *JsonSchema, str *JsonStruct) bool {
	if r.Intn(10) == 0 {
		field := JsonStructField{
			JsonFieldType: JsonFieldType{},
			Name:          fmt.Sprintf("Field#%d", len(str.Fields)+1),
		}

		p := PrimitiveTypeString
		switch r.Intn(4) {
		case 0:
			field.JsonFieldType.Primitive = &p
			if r.Intn(10) == 0 {
				field.DictName = "Dict" + field.Name
			}

		case 1:
			f := JsonFieldType{Primitive: &p}
			field.JsonFieldType.Array = &f
		case 2:
			multimapIdx := r.Intn(len(schema.Multimaps))
			i := 0
			for multimapName := range schema.Multimaps {
				if i == multimapIdx {
					field.JsonFieldType.MultiMap = multimapName
					break
				}
				i++
			}
		case 3:
			if r.Intn(2) == 0 {
				// Add new struct
				struc := JsonStruct{
					Name:   fmt.Sprintf("Struct#%d", len(schema.Structs)),
					Fields: []JsonStructField{},
				}
				schema.Structs[struc.Name] = &struc
				field.JsonFieldType.Struct = struc.Name
			} else {
				structIdx := r.Intn(len(schema.Structs))
				i := 0
				for structName := range schema.Structs {
					if i == structIdx {
						field.JsonFieldType.Struct = structName
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

func shrinkSchema(t *testing.T, r *rand.Rand, orig *JsonSchema) (cpy *JsonSchema) {
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

func shrinkStruct(t *testing.T, r *rand.Rand, schema *JsonSchema, str *JsonStruct) bool {
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

	orig := &JsonSchema{}
	err = json.Unmarshal(schemaJson, &orig)
	require.NoError(t, err)
	orig, err = orig.PrunedForRoot(orig.MainStruct)
	require.NoError(t, err)

	r := rand.New(rand.NewSource(42))

	// Expand one field at a time and check compatibility.
	for i := 0; i < 200; i++ {
		expanded := expandSchema(t, r, orig)
		expandedWire, err := expanded.ToWire()
		require.NoError(t, err)

		// Exact compatible with itself
		compat, err := expandedWire.Compatible(expandedWire)
		require.NoError(t, err, i)
		assert.EqualValues(t, CompatibilityExact, compat, i)

		// Expanding is compatible / superset
		origWire, err := orig.ToWire()
		require.NoError(t, err, i)
		compat, err = expandedWire.Compatible(origWire)
		require.NoError(t, err, i)
		assert.EqualValues(t, CompatibilitySuperset, compat, i)

		// Opposite direction is incompatible
		compat, err = origWire.Compatible(expandedWire)
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
		require.EqualValues(t, expandedWire, &cpy)

		orig = expanded
	}
}

func TestSchemaShrink(t *testing.T) {
	schemaJson, err := os.ReadFile("testdata/oteltef.wire.json")
	require.NoError(t, err)

	orig := &JsonSchema{}
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
		shrinkedWire, err := shrinked.ToWire()
		require.NoError(t, err)

		// Shrinking is incompatible
		origWire, err := orig.ToWire()
		require.NoError(t, err, i)
		compat, err := shrinkedWire.Compatible(origWire)
		require.Error(t, err, i)
		assert.EqualValues(t, CompatibilityIncompatible, compat, i)

		// Opposite direction is compatible/superset
		compat, err = origWire.Compatible(shrinkedWire)
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
		require.EqualValues(t, shrinkedWire, &cpy)

		orig = shrinked
	}
}
