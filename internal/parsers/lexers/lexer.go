// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package lexers implements a lexer for TribeNet turn reports
package lexers

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Lexer struct {
	line, col int // position in the source

	pos int  // offset of the next byte in the source
	ch  rune // current character

	// the source is owned by the caller and must not be altered
	input []byte
}

func New(line, col int, input []byte) *Lexer {
	return &Lexer{
		line:  line,
		col:   col,
		input: input,
	}
}

// Next returns the next token in the source.
// Returns nil if we've reached the end of input before we start scanning.
//
// Returns a token with kind == EOF if we reach end of input while scanning.
// This allows us to collect any trivia after the previous token.
func (l *Lexer) Next() *Token {
	if l.isEOF() {
		return nil
	}

	var t Token

	// collect leading trivia.
	// right now, this is whitespace (not including end-of-line).
	for !l.isEOF() {
		if l.isWhitespace() {
			trivia := Trivia{
				Kind: Whitespace,
				Span: Span{Start: l.pos, Line: l.line, Col: l.col},
			}
			l.skipWhitespace()
			trivia.Span.End = l.pos
			t.LeadingTrivia = append(t.LeadingTrivia, trivia)
			continue
		}
		break
	}

	// collect the token
	t.Span = Span{Start: l.pos, Line: l.line, Col: l.col}
	if l.isEOF() {
		t.Kind = EOF
	} else if ch := l.current(); ch == '\n' {
		t.Kind = EOL
		l.advance()
	} else if ch == ',' {
		t.Kind = Comma
		l.advance()
	} else if ch == '=' {
		t.Kind = Equal
		l.advance()
	} else if ch == '#' {
		t.Kind = Hash
		l.advance()
	} else if ch == '(' {
		t.Kind = LParen
		l.advance()
	} else if ch == ')' {
		t.Kind = RParen
		l.advance()
	} else if l.isDigit() {
		t.Kind = Number
		// advance past the digits
		for l.isDigit() {
			l.advance()
		}
	} else if l.isAlpha() {
		t.Kind = Text
		// advance past the text
		for l.isText() {
			l.advance()
		}
	} else {
		// unrecognized input is collected until we find something that we recognize
		t.Kind = Unknown
		l.skipUnknown()
	}
	t.Span.End = l.pos

	// return the token
	return &t
}

// Text is a helper for diagnostics / debugging.
func (l *Lexer) Text(t *Token) string {
	if t == nil {
		panic("assert(token != nil)")
	}
	return t.Span.Text(l.input)
}

// TextWithTrivia is a helper for diagnostics / debugging.
func (l *Lexer) TextWithTrivia(t *Token) string {
	if t == nil {
		panic("assert(token != nil)")
	}
	return t.TextWithTrivia(l.input)
}

const (
	eofRune rune = -1
)

// advance moves the position to the next character if we're not at the end of input.
func (l *Lexer) advance() {
	// eof fast path
	if l.pos >= len(l.input) {
		return
	}
	// we have to decode current rune to figure out how to advance the column and line numbers
	r, w := utf8.DecodeRune(l.input[l.pos:])
	l.pos += w
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
}

// current returns the current character in the input.
func (l *Lexer) current() rune {
	if l.pos >= len(l.input) {
		return eofRune
	}
	// decode current rune. the CST that drives the lexers will handle RuneError when needed.
	r, _ := utf8.DecodeRune(l.input[l.pos:])
	return r
}

// isAlpha returns true if the character is a letter
func (l *Lexer) isAlpha() bool {
	return unicode.IsLetter(l.current())
}

// isDigit returns true if the character is a digit
func (l *Lexer) isDigit() bool {
	return unicode.IsDigit(l.current())
}

// isEOF returns true at end of input
func (l *Lexer) isEOF() bool {
	return l.pos >= len(l.input)
}

// isEOL returns true if the character is an end-of-line
func (l *Lexer) isEOL() bool {
	return l.current() == '\n'
}

// isPunctuation returns true if the character is punctuation
func (l *Lexer) isPunctuation() bool {
	// we have to check for eof because we're going to convert the
	// current rune to a byte for the punctuation check.
	if l.isEOF() {
		return false
	}
	ch := byte(l.current())
	return bytes.IndexByte([]byte{',', '=', '#', '(', ')'}, ch) != -1
}

// isText returns true if the char is text
func (l *Lexer) isText() bool {
	ch := l.current()
	return unicode.IsLetter(ch) || unicode.IsDigit(ch)
}

// isWhitespace returns true if the character is space or tab
func (l *Lexer) isWhitespace() bool {
	ch := l.current()
	return ch != '\n' && unicode.IsSpace(ch)
}

// skipUnknown skips characters that aren't known to be valid in a report
func (l *Lexer) skipUnknown() {
	for !(l.isEOF() || l.isDigit() || l.isPunctuation() || l.isText()) {
		l.advance()
	}
}

// skipWhitespace skips spaces and tabs
func (l *Lexer) skipWhitespace() {
	for l.isWhitespace() {
		l.advance()
	}
}

type Span struct {
	Start int // byte offset (inclusive)
	End   int // byte offset (exclusive)
	Line  int // 1-based
	Col   int // 1-based, in UTF-8 code points
}

// Text is a helper for diagnostics / debugging.
func (s Span) Text(src []byte) string {
	return string(src[s.Start:s.End])
}

type Trivia struct {
	Kind TriviaKind
	Span Span
}

// Text is a helper for diagnostics / debugging.
func (t Trivia) Text(src []byte) string {
	return t.Span.Text(src)
}

type TriviaKind int

const (
	InvalidRunes TriviaKind = iota
	Comment
	Whitespace
)

func (k TriviaKind) String() string {
	switch k {
	case InvalidRunes:
		return "InvalidRunes"
	case Comment:
		return "Comment"
	case Whitespace:
		return "Whitespace"
	default:
		return fmt.Sprintf("TriviaKind(%d)", k)
	}
}

type Token struct {
	Kind           TokenKind
	Span           Span
	LeadingTrivia  []Trivia
	TrailingTrivia []Trivia
}

// Text is a helper for diagnostics / debugging.
func (t *Token) Text(src []byte) string {
	if t == nil {
		panic("assert(token != nil)")
	}
	return t.Span.Text(src)
}

// TextWithTrivia is a helper for diagnostics / debugging.
func (t *Token) TextWithTrivia(src []byte) string {
	if t == nil {
		panic("assert(token != nil)")
	}
	var sb strings.Builder
	for _, trivia := range t.LeadingTrivia {
		sb.WriteString(trivia.Text(src))
	}
	sb.WriteString(t.Span.Text(src))
	for _, trivia := range t.TrailingTrivia {
		sb.WriteString(trivia.Text(src))
	}
	return sb.String()
}

type TokenKind int

const (
	EOF TokenKind = iota
	EOL

	// for now, let's only have text, numeric, and punctuation for our tokens.
	// we will update this when we have a better feel for how the lexer is implemented.

	Text

	Number

	Comma
	Equal
	Hash
	LParen
	RParen

	Unknown
)

func (tk TokenKind) String() string {
	switch tk {
	case EOF:
		return "EOF"
	case EOL:
		return "EOL"
	case Text:
		return "Text"
	case Number:
		return "Number"
	case Comma:
		return "Comma"
	case Equal:
		return "Equal"
	case Hash:
		return "Hash"
	case LParen:
		return "LParen"
	case RParen:
		return "RParen"
	case Unknown:
		return "Unknown"
	default:
		return fmt.Sprintf("TokenKind(%d)", tk)
	}
}
