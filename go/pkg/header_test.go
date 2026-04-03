package pkg

import (
	"bytes"
	"encoding/binary"
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

// TestVarHeaderDeserializeExploitTruncatedStringAccepted demonstrates a vulnerability:
// Deserialize accepts a string whose declared length exceeds available bytes, and
// returns a zero-padded value instead of an error.
func TestVarHeaderDeserializeExploitTruncatedStringAccepted(t *testing.T) {
	orig := VarHeader{
		UserData: map[string]string{"k": "v"},
	}
	var serialized bytes.Buffer
	err := orig.Serialize(&serialized)
	require.NoError(t, err)

	payload := append([]byte(nil), serialized.Bytes()...)
	mutateVarHeaderUserDataValueLen(t, payload, 5)

	var decoded VarHeader
	err = decoded.Deserialize(bytes.NewBuffer(payload))
	require.Error(t, err, "this was passing before the vulnerability was fixed")
}

func mutateVarHeaderUserDataValueLen(t *testing.T, payload []byte, newLen byte) {
	t.Helper()

	p := 0

	schemaLen, n := binary.Uvarint(payload[p:])
	require.Greater(t, n, 0)
	p += n + int(schemaLen)

	count, n := binary.Uvarint(payload[p:])
	require.Greater(t, n, 0)
	require.Equal(t, uint64(1), count)
	p += n

	kLen, n := binary.Uvarint(payload[p:])
	require.Greater(t, n, 0)
	p += n + int(kLen)

	vLen, n := binary.Uvarint(payload[p:])
	require.Greater(t, n, 0)
	require.Equal(t, uint64(1), vLen)
	payload[p] = newLen
}
