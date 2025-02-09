package idl

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexer(t *testing.T) {
	l := NewLexer(bytes.NewBufferString("struct abc {}"))
	l.Start()

	tokens := []Token{tStruct, tIdent, tLBrace, tRBrace, tEOF}
	i := 0
	for {
		l.Next()
		token := l.Token()
		assert.Equal(t, tokens[i], token, i)
		i++
		if token == tEOF {
			break
		}
	}
}
