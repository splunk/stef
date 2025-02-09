package idl

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLexer(t *testing.T) {
	l := NewLexer(bytes.NewBufferString("struct abc {}"))

	tokens := []Token{tStruct, tIdent, tLBrace, tRBrace, tEOF}
	i := 0
	for {
		token := l.Token()
		assert.Equal(t, tokens[i], token, i)
		i++
		if token == tEOF {
			break
		}
		l.Next()
	}
}

func FuzzLexer(f *testing.F) {
	f.Add([]byte(nil))
	f.Add([]byte(""))
	f.Add([]byte("struct abc {}"))

	testFiles := []string{"testdata/example.stef", "testdata/otel.stef"}
	for _, file := range testFiles {
		content, err := os.ReadFile(file)
		require.NoError(f, err)
		f.Add(content)
	}

	f.Fuzz(
		func(t *testing.T, content []byte) {
			l := NewLexer(bytes.NewBuffer(content))
			for {
				token := l.Token()
				if token == tEOF || token == tError {
					break
				}
				l.Next()
			}
		},
	)
}
