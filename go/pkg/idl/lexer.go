package idl

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"unicode"
)

// Lexer splits a UTF8-encoded input into tokens.
type Lexer struct {
	input *bufio.Reader
	token Token

	nextRune  rune
	prevWasCR bool

	isEOF   bool
	isError bool
	errMsg  string

	curPos  Pos
	prevPos Pos

	identRunes []rune
	ident      string
}

// Pos indicates a position in the input stream.
type Pos struct {
	ByteOfs uint
	Line    uint
	Col     uint
}

type Token uint

const (
	tError Token = iota
	tEOF

	tPackage
	tIdent

	tStruct
	tOneof
	tMultimap

	tOptional
	tRoot
	tDict
	tKey
	tValue

	tBool
	tInt64
	tUint64
	tFloat64
	tString
	tBytes

	tLBracket = '['
	tRBracket = ']'
	tLParen   = '('
	tRParen   = ')'
	tLBrace   = '{'
	tRBrace   = '}'
)

func (t Token) String() string {
	str, ok := keywordsReverse[t]
	if ok {
		return str
	}
	switch t {
	case tEOF:
		return "EOF"
	case tIdent:
		return "identifier"
	default:
		return string(byte(t))
	}
}

var keywords = map[string]Token{
	"package":  tPackage,
	"struct":   tStruct,
	"oneof":    tOneof,
	"multimap": tMultimap,
	"optional": tOptional,
	"root":     tRoot,
	"dict":     tDict,
	"key":      tKey,
	"value":    tValue,
	"bool":     tBool,
	"int64":    tInt64,
	"uint64":   tUint64,
	"float64":  tFloat64,
	"string":   tString,
	"bytes":    tBytes,
}

var keywordsReverse = func() map[Token]string {
	m := make(map[Token]string)
	for k, v := range keywords {
		m[v] = k
	}
	return m
}()

func NewLexer(input io.Reader) *Lexer {
	l := &Lexer{
		input: bufio.NewReader(input),
		curPos: Pos{
			ByteOfs: 0,
			Line:    1,
			Col:     1,
		},
	}
	// Fetch the first rune.
	l.readNextRune()

	// Fetch the first token.
	l.Next()
	return l
}

func (l *Lexer) Token() Token {
	return l.token
}

// Next reads the input for the next token. After that Token() will return
// the token that was read.
//
// If Lexer input is at EOF then the next Token() call will return tEOF.
// If reading failed (e.g. if the input is not valid UTF8) the next Token() call
// will return tError.
func (l *Lexer) Next() {
	l.prevPos = l.curPos

	l.skipWhiteSpaceOrComment()

	if l.isEOF {
		l.token = tEOF
		return
	} else if l.isError {
		l.token = tError
		l.isError = false
		return
	}

	switch l.nextRune {
	case tLParen:
		l.token = tLParen
	case tRParen:
		l.token = tRParen
	case tLBracket:
		l.token = tLBracket
	case tRBracket:
		l.token = tRBracket
	case tRBrace:
		l.token = tRBrace
	case tLBrace:
		l.token = tLBrace
	default:
		if unicode.IsLetter(l.nextRune) {
			// This is a letter. It must a start of an identifier or keyword.
			l.readIdentOrKeyword()
			return
		}
		l.token = tError
		l.errMsg = fmt.Sprintf("invalid character: %c", l.nextRune)
	}
	l.readNextRune()
}

func (l *Lexer) skipWhiteSpaceOrComment() {
	for !l.isEOF && !l.isError {
		if unicode.IsSpace(l.nextRune) {
			l.readNextRune()
		} else if l.nextRune == '/' {
			l.skipComment()
		} else {
			break
		}
	}
}

func (l *Lexer) skipComment() {
	l.readNextRune()
	if l.isEOF || l.isError || l.nextRune != '/' {
		l.token = tError
		l.errMsg = "expected start of comment"
		return
	}

	for !l.isEOF && !l.isError && l.nextRune != '\r' && l.nextRune != '\n' {
		l.readNextRune()
	}
}

func (l *Lexer) readNextRune() {
	nextRune, size, err := l.input.ReadRune()
	if err != nil {
		if errors.Is(err, io.EOF) {
			l.isEOF = true
		} else {
			l.isError = true
			l.errMsg = fmt.Sprintf("invalid character")
		}
		return
	}
	l.nextRune = nextRune
	l.curPos.ByteOfs += uint(size)
	l.curPos.Col++

	// Handle any of CR,LF,CRLF as a new line.
	const cCR = '\r'
	const cLF = '\n'
	if l.nextRune == cCR {
		l.curPos.Line++
		l.curPos.Col = 1
		l.prevWasCR = true
	} else if l.nextRune == cLF {
		if !l.prevWasCR {
			l.curPos.Line++
			l.curPos.Col = 1
		}
		l.prevWasCR = false
	} else {
		l.prevWasCR = false
	}
}

func (l *Lexer) readIdentOrKeyword() Token {
	l.identRunes = l.identRunes[:0]

	// The first character is already read. Subsequent characters must be
	// letters, digits or underscore.
	for (unicode.IsLetter(l.nextRune) || unicode.IsDigit(l.nextRune) || l.nextRune == '_') && !l.isError {
		l.identRunes = append(l.identRunes, l.nextRune)
		l.readNextRune()
		if l.isEOF {
			break
		}
		if l.isError {
			l.token = tError
			return tError
		}
	}

	l.ident = string(l.identRunes)

	// Check if it is a keyword.
	if token, ok := keywords[l.ident]; ok {
		l.token = token
		return token
	}

	l.token = tIdent
	return tIdent
}

// Ident will return the identifier if the current token is tIdent.
// Use Token() first.
func (l *Lexer) Ident() string {
	return l.ident
}

// TokenStartPos will return the starting position of the last read
// token after Next() call.
func (l *Lexer) TokenStartPos() Pos {
	return l.prevPos
}
