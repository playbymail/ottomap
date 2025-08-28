// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package lexers implements a lexer for TribeNet turn reports
package lexers

type Lexer struct {
}

func New(line, col int, input []byte) *Lexer {
	return &Lexer{}
}

func (l *Lexer) Next() *Token {
	return nil
}

type Span struct {
	Start int // byte offset (inclusive)
	End   int // byte offset (exclusive)
	Line  int // 1-based
	Col   int // 1-based, in UTF-8 code points
}

type Trivia struct {
	Kind TriviaKind // Whitespace, LineComment, BlockComment
	Span Span
}

type TriviaKind int

func (k TriviaKind) String() string {
	return "!implemented"
}

const (
	Whitespace TriviaKind = iota
	LineComment
	BlockComment
)

type Token struct {
	Kind           TokenKind
	Span           Span
	LeadingTrivia  []Trivia
	TrailingTrivia []Trivia
}

type TokenKind int

const (
	EOF TokenKind = iota
	KeywordCurrent
	KeywordTurn
	MonthName
	Identifier
	Number
	LParen
	Hash
	RParen
)

func (tk TokenKind) String() string {
	return "!implemented"
}

// Text is a helper for diagnostics / debugging:
func (t Token) Text(src []byte) string {
	return string(src[t.Span.Start:t.Span.End])
}
