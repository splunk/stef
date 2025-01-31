package pkg

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var varHeadTests = []VarHeader{
	{},
	{
		SchemaWireBytes: []byte{},
		UserData:        map[string]string{},
	},
	{
		SchemaWireBytes: []byte("012"),
		UserData:        map[string]string{},
	},
	{
		SchemaWireBytes: []byte("012345"),
		UserData:        map[string]string{"abc": "def", "0": "world"},
	},
}

func TestVarHeaderSerialization(t *testing.T) {
	for _, orig := range varHeadTests {
		var buf bytes.Buffer
		err := orig.Serialize(&buf)
		require.NoError(t, err)

		var cpy VarHeader
		err = cpy.Deserialize(&buf)
		require.NoError(t, err)

		if len(orig.SchemaWireBytes) == 0 {
			assert.True(t, len(cpy.SchemaWireBytes) == 0)
		} else {
			assert.EqualValues(t, orig.SchemaWireBytes, cpy.SchemaWireBytes)
		}
		if len(orig.UserData) == 0 {
			assert.True(t, len(cpy.UserData) == 0)
		} else {
			assert.EqualValues(t, orig.UserData, cpy.UserData)
		}
	}
}

func FuzzVarHeaderDeserialize(f *testing.F) {
	for _, hdr := range varHeadTests {
		var buf bytes.Buffer
		err := hdr.Serialize(&buf)
		require.NoError(f, err)
		f.Add(buf.Bytes())
	}

	f.Fuzz(
		func(t *testing.T, data []byte) {
			var hdr VarHeader
			_ = hdr.Deserialize(bytes.NewBuffer(data))
		},
	)
}
