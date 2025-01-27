package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
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
