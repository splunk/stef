package idl

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/splunk/stef/go/pkg/schema"
)

func TestParser(t *testing.T) {
	idlBytes, err := os.ReadFile("testdata/otel.stef")
	require.NoError(t, err)

	lexer := NewLexer(bytes.NewBuffer(idlBytes))
	parser := NewParser(lexer)
	err = parser.Parse()
	require.NoError(t, err)

	jsonBytes, err := os.ReadFile("testdata/oteltef.wire.json")
	require.NoError(t, err)

	var schem schema.Schema
	err = json.Unmarshal(jsonBytes, &schem)
	require.NoError(t, err)

	require.EqualValues(t, &schem, parser.Schema())
}
