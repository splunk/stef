package schema

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJsonToWire(t *testing.T) {
	jsonBytes, err := os.ReadFile("testdata/example.json")
	require.NoError(t, err)

	var jsonSchema JsonSchema
	err = json.Unmarshal(jsonBytes, &jsonSchema)
	require.NoError(t, err)

	wireSchema, err := jsonSchema.ToWire()
	require.NoError(t, err)
	require.NotNil(t, wireSchema)
}
