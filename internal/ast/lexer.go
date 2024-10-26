// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package ast

import (
	"unicode/utf8"
)

func NewLexer(input []byte) *Lexer {
	l := &Lexer{
		input: input,
		pos:   0,
		line:  1,
		col:   1,
	}
	return l
}

type Lexer struct {
	input []byte // the original input buffer (not a copy)
	pos   int    // position of the next character to be read
	line  int
	col   int
}

// Next returns the next token in the input.
// If the end of the input is reached, it returns an EOF token.
func (l *Lexer) Next() Token {
	tok := Token{
		Line: l.line,
		Col:  l.col,
	}

	if l.pos >= len(l.input) {
		tok.Type = EOF
		return tok
	}

	// we are going to chunk the input into four categories:
	// 1. run of invalid utf8 (type will be INVALID_UTF8)
	// 3. newline (type will be EOL)
	// 2. run of whitespace (type will be SPACES)
	// 4. words (type will be WORD)
	start := l.pos
	r, w := utf8.DecodeRune(l.input[start:])
	l.pos, l.col = l.pos+w, l.col+1

	if r == utf8.RuneError {
		// invalid utf8 sequence means we must advance the input to the the first valid rune
		for l.pos < len(l.input) {
			if r, w = utf8.DecodeRune(l.input[l.pos:]); r != utf8.RuneError {
				break
			}
			l.pos, l.col = l.pos+w, l.col+1
		}
		tok.Type = INVALID_UTF8
		return tok
	}

	return tok
}

// Token is a token in the input.
// The Literal field is allocated by the lexer so the caller can free the original input.
type Token struct {
	Type      TokenType
	Line, Col int
	Literal   []byte // the literal value of the token. this is allocated by the lexer.
}

type TokenType int

const (
	EOF TokenType = iota
	EOL
	INVALID_UTF8
	SPACES
	WORD
)
