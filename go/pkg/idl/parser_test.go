package idl

import (
	"bytes"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParserErrors(t *testing.T) {
	tests := []struct {
		input string
		err   string
	}{
		{
			input: "package ",
			err:   "test.stef:1:9: identifier expected",
		},
		{
			input: "package abc.",
			err:   "test.stef:1:13: identifier expected",
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
		{
			input: "package abc\nstruct MyStruct {\nField UnknownType }",
			err:   "test.stef:3:20: unknown type: UnknownType",
		},
		{
			input: "package abc oneof A {} struct A {}",
			err:   "test.stef:1:31: duplicate top-level identifier: A",
		},
		{
			input: "package abc enum {}",
			err:   "test.stef:1:18: enum name expected",
		},
		{
			input: "package abc enum Enum { Value = }",
			err:   "test.stef:1:33: enum field value expected",
		},
		{
			input: "struct Root root {}",
			err:   "test.stef:1:2: expected package but got struct",
		},
		{
			input: "package abc struct Root root { x1 int64 x1 string }",
			err:   "test.stef:1:41: duplicate field name: x1",
		},
		{
			input: "package abc struct Root root { x1 int64 dict(DictName) }",
			err:   "test.stef:1:41: only string or bytes can have dict modifier",
		},
		{
			input: "package abc oneof Root root {}",
			err:   "test.stef:1:24: oneof cannot be a root",
		},
		{
			input: "package abc oneof Oneof dict(abc) {}",
			err:   "test.stef:1:25: oneof cannot have dict modifier",
		},
		{
			input: "package abc struct Empty root {}",
			err:   "test.stef:1:33: root struct must have at least one field",
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

func TestParserWarnings(t *testing.T) {
	tests := []struct {
		input string
		warn  []string
	}{
		{
			input: "package abc struct R root { x bool } struct U {}",
			warn:  []string{`Warning: struct "U" is defined but not used`},
		},
		{
			input: "package abc struct R root { x bool } multimap M { key string value int64 }",
			warn:  []string{`Warning: multimap "M" is defined but not used`},
		},
		{
			input: "package abc struct R root { x bool } oneof O {}",
			warn:  []string{`Warning: oneof "O" is defined but not used`},
		},
		{
			input: "package abc struct R root { x bool } enum E {}",
			warn:  []string{`Warning: enum "E" is defined but not used`},
		},
		{
			input: "package abc struct R root { x bool } struct A { B B optional } struct B { A A optional }",
			warn: []string{
				`Warning: struct "A" is defined but not used`, `Warning: struct "B" is defined but not used`,
			},
		},
		// Multiple unused types of different kinds
		{
			input: "package abc struct R root { x bool } struct U1 {} struct U2 {} oneof O {} multimap M { key string value int64 } enum E {}",
			warn: []string{
				`Warning: oneof "O" is defined but not used`,
				`Warning: struct "U1" is defined but not used`,
				`Warning: struct "U2" is defined but not used`,
				`Warning: multimap "M" is defined but not used`,
				`Warning: enum "E" is defined but not used`,
			},
		},
		// Chain of unused structs where one references another
		{
			input: "package abc struct R root { x bool } struct A { b B } struct B { c C } struct C {}",
			warn: []string{
				`Warning: struct "A" is defined but not used`,
				`Warning: struct "B" is defined but not used`,
				`Warning: struct "C" is defined but not used`,
			},
		},
		// Circular dependency that's unused
		{
			input: "package abc struct R root { x bool } struct A { b []B c C } struct B { a A } struct C { name string }",
			warn: []string{
				`Warning: struct "A" is defined but not used`,
				`Warning: struct "B" is defined but not used`,
				`Warning: struct "C" is defined but not used`,
			},
		},
		// Mixed used and unused with nested dependencies
		{
			input: "package abc struct R root { data Data } struct Data { items []Item } struct Item {} struct UnusedHelper { V string } oneof UnusedChoice {}",
			warn: []string{
				`Warning: oneof "UnusedChoice" is defined but not used`,
				`Warning: struct "UnusedHelper" is defined but not used`,
			},
		},
		// Only root struct, no warnings
		{
			input: "package abc struct R root { name string }",
			warn:  []string{},
		},
		// Oneof used in struct field
		{
			input: "package abc struct R root { choice Choice } oneof Choice { option1 Option1 option2 Option2 } struct Option1 {} struct Option2 {} struct Unused {}",
			warn: []string{
				`Warning: struct "Unused" is defined but not used`,
			},
		},
		// Deep nesting with some unused branches
		{
			input: "package abc struct R root { level1 Level1 } struct Level1 { level2 Level2 } struct Level2 { V string } struct UnusedBranch { deep UnusedDeep } struct UnusedDeep {}",
			warn: []string{
				`Warning: struct "UnusedBranch" is defined but not used`,
				`Warning: struct "UnusedDeep" is defined but not used`,
			},
		},
		// Optional field with unused type
		{
			input: "package abc struct R root { opt OptionalStruct optional } struct OptionalStruct {} struct Unused {}",
			warn: []string{
				`Warning: struct "Unused" is defined but not used`,
			},
		},
		// Array field with unused other types
		{
			input: "package abc struct R root { items []ItemType } struct ItemType {} struct UnrelatedType {} enum UnrelatedEnum {}",
			warn: []string{
				`Warning: struct "UnrelatedType" is defined but not used`,
				`Warning: enum "UnrelatedEnum" is defined but not used`,
			},
		},
		// Complex mix of all type kinds with various usage patterns
		{
			input: "package abc struct R root { used UsedStruct choice UsedOneof tags UsedMap } struct UsedStruct { status UsedEnum } oneof UsedOneof { opt1 Option1 opt2 Option2 } struct Option1 {} struct Option2 {} multimap UsedMap { key string value MapValue } struct MapValue {} enum UsedEnum { A = 0 B = 1 } struct Unused1 {} struct Unused2 {} oneof UnusedOneof {} multimap UnusedMap { key string value string } enum UnusedEnum {}",
			warn: []string{
				`Warning: struct "Unused1" is defined but not used`,
				`Warning: struct "Unused2" is defined but not used`,
				`Warning: oneof "UnusedOneof" is defined but not used`,
				`Warning: multimap "UnusedMap" is defined but not used`,
				`Warning: enum "UnusedEnum" is defined but not used`,
			},
		},
		// Self-referencing unused struct
		{
			input: "package abc struct R root { x bool } struct SelfRef { next SelfRef optional }",
			warn: []string{
				`Warning: struct "SelfRef" is defined but not used`,
			},
		},
		// Multiple levels of mutual references that are unused
		{
			input: "package abc struct R root { x bool } struct A { b B } struct B { c C } struct C { a A optional }",
			warn: []string{
				`Warning: struct "A" is defined but not used`,
				`Warning: struct "B" is defined but not used`,
				`Warning: struct "C" is defined but not used`,
			},
		},
	}

	for i, test := range tests {
		t.Run(
			strconv.Itoa(i), func(t *testing.T) {
				lexer := NewLexer(bytes.NewBufferString(test.input))
				parser := NewParser(lexer, "test.stef")
				err := parser.Parse()
				require.NoError(t, err)
				msgs := parser.Messages()
				require.Equal(t, len(test.warn), len(msgs))
				for i := range test.warn {
					require.Equal(t, test.warn[i], msgs[i].String())
				}
			},
		)
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
