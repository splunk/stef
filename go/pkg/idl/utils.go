package idl

import (
	"bytes"

	"github.com/splunk/stef/go/pkg/schema"
)

// Parse parses the schema from the provided STEF IDL.
func Parse(schemaIDL []byte, fileName string) (*schema.Schema, error) {
	lexer := NewLexer(bytes.NewBuffer(schemaIDL))
	parser := NewParser(lexer, fileName)
	err := parser.Parse()
	return parser.Schema(), err
}
