package idl

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/splunk/stef/go/pkg/schema"
)

func TestParserErrors(t *testing.T) {
	tests := []struct {
		input string
		err   string
	}{
		{
			input: "package ",
			err:   "test.stef:1:9: package name expected",
		},
		{
			input: "package abc\nhello",
			err:   "test.stef:2:1: expected struct, oneof or multimap",
		},
		{
			input: "package abc\nstruct string",
			err:   "test.stef:2:8: struct name expected",
		},
		{
			input: "package abc\nmultimap [",
			err:   "test.stef:2:10: multimap name expected",
		},
		{
			input: "package abc\nstruct MyStruct dict()",
			err:   "test.stef:2:23: dict name expected",
		},
		{
			input: "package abc\nstruct MyStruct dict[]",
			err:   "test.stef:2:22: expected ( but got [",
		},
		{
			input: "package abc\nstruct MyStruct {\nField []struct",
			err:   "test.stef:3:10: type specifier expected after []",
		},
	}

	for _, test := range tests {
		lexer := NewLexer(bytes.NewBufferString(test.input))
		parser := NewParser(lexer, "test.stef")
		err := parser.Parse()
		require.Error(t, err)
		require.Equal(t, test.err, err.Error())
	}
}

func TestParseExample(t *testing.T) {
	inputFile := "testdata/example.stef"
	idlBytes, err := os.ReadFile(inputFile)
	require.NoError(t, err)

	lexer := NewLexer(bytes.NewBuffer(idlBytes))
	parser := NewParser(lexer, inputFile)
	err = parser.Parse()
	require.NoError(t, err)
}

func TestParserOtelSTEF(t *testing.T) {
	inputFile := "testdata/otel.stef"
	idlBytes, err := os.ReadFile(inputFile)
	require.NoError(t, err)

	lexer := NewLexer(bytes.NewBuffer(idlBytes))
	parser := NewParser(lexer, inputFile)
	err = parser.Parse()
	require.NoError(t, err)

	jsonBytes, err := os.ReadFile("testdata/oteltef.wire.json")
	require.NoError(t, err)

	var schem schema.Schema
	err = json.Unmarshal(jsonBytes, &schem)
	require.NoError(t, err)

	require.EqualValues(t, &schem, parser.Schema())
}

func FuzzParser(f *testing.F) {
	testFiles := []string{"testdata/example.stef", "testdata/otel.stef"}
	for _, file := range testFiles {
		content, err := os.ReadFile(file)
		require.NoError(f, err)
		f.Add(content)
	}

	f.Fuzz(
		func(t *testing.T, content []byte) {
			p := NewParser(NewLexer(bytes.NewBuffer(content)), "temp.stef")
			_ = p.Parse()
		},
	)
}
