// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package cst implements a lossless CST parser over tokens produced by the lexers package.
package cst

import (
	"fmt"

	"github.com/playbymail/ottomap/internal/reports/lexers"
)

// ====== Public API ======

// ParseFile parses one input buffer into a lossless CST. It never panics.
// It returns a CST (possibly with Bad* nodes) and any diagnostics encountered.
func ParseFile(input []byte) (*File, []Diagnostic) {
	p := &Parser{
		lx:    lexers.New(1, 1, input),
		input: input,
	}
	p.la = p.lx.Next()

	f := &File{}
	for p.la != nil {
		if p.at(lexers.KeywordCurrent) {
			h := p.parseHeader()
			f.Decls = append(f.Decls, h)
			continue
		}
		// Unknown top-level construct. Capture raw tokens until a sync point.
		bad := p.parseBadTopLevel()
		f.Decls = append(f.Decls, bad)
	}
	f.EOF = p.synthesizeEOF()
	f.span = coverNodesAndTokens(f.Decls, f.EOF)
	return f, p.diags
}

// ====== Diagnostics ======

type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
	SeverityInfo
)

type Diagnostic struct {
	Severity Severity
	Span     Span
	Message  string
	Notes    []string
}

func (p *Parser) addError(span Span, msg string, notes ...string) {
	p.diags = append(p.diags, Diagnostic{
		Severity: SeverityError,
		Span:     span,
		Message:  msg,
		Notes:    append([]string(nil), notes...),
	})
}

// ====== Spans ======

// Span is the CST span type. We keep it separate from lexers.Span to avoid leaking internals.
type Span struct {
	Start int // byte offset (inclusive)
	End   int // byte offset (exclusive)
	Line  int // 1-based
	Col   int // 1-based (UTF-8 code points)
}

func convertSpan(s lexers.Span) Span {
	return Span{Start: s.Start, End: s.End, Line: s.Line, Col: s.Col}
}

func convertBack(s Span) lexers.Span {
	return lexers.Span{Start: s.Start, End: s.End, Line: s.Line, Col: s.Col}
}

// insertionSpan returns a zero-width span at the insertion point for synthesized tokens.
func (p *Parser) insertionSpan(anchor *lexers.Token) Span {
	if anchor != nil {
		as := convertSpan(anchor.Span)
		// Insert "before" the anchor token at its start.
		return Span{Start: as.Start, End: as.Start, Line: as.Line, Col: as.Col}
	}
	// No lookahead: use the end of last consumed span (or 0,0).
	if p.last.End > 0 {
		return Span{Start: p.last.End, End: p.last.End, Line: p.last.Line, Col: p.last.Col}
	}
	return Span{} // zero; file start
}

// ====== Node Kinds & Interfaces ======

type Kind int

const (
	KindFile Kind = iota
	KindHeader
	KindBadTopLevel
	KindToken
)

func (k Kind) String() string {
	return "!implemented"
}

type Node interface {
	Span() Span
	Kind() Kind
}

// ====== TokenNode (wraps a lexer token) ======

type TokenNode struct {
	Tok *lexers.Token // carries Kind, Span, LeadingTrivia, TrailingTrivia
}

func (t *TokenNode) Span() Span { return convertSpan(t.Tok.Span) }
func (t *TokenNode) Kind() Kind { return KindToken }

// ====== CST Root ======

type File struct {
	Decls []Node
	EOF   *TokenNode
	span  Span
}

func (f *File) Span() Span { return f.span }
func (f *File) Kind() Kind { return KindFile }

// ====== Header Node ======

type Header struct {
	KwCurrent *TokenNode // KeywordCurrent
	KwTurn    *TokenNode // KeywordTurn
	Month     *TokenNode // MonthName | Identifier
	Year      *TokenNode // Number
	LParen    *TokenNode // "("
	Hash      *TokenNode // "#"
	TurnNo    *TokenNode // Number
	RParen    *TokenNode // ")"
	span      Span
}

func (h *Header) Span() Span { return h.span }
func (h *Header) Kind() Kind { return KindHeader }

// ====== BadTopLevel Node ======

type BadTopLevel struct {
	Tokens []*TokenNode
	span   Span
}

func (b *BadTopLevel) Span() Span { return b.span }
func (b *BadTopLevel) Kind() Kind { return KindBadTopLevel }

// ====== Parser State ======

type Parser struct {
	lx    *lexers.Lexer
	la    *lexers.Token // lookahead; nil only at EOF
	input []byte

	diags []Diagnostic

	// Track last consumed token span (for EOF insertion points etc.).
	last Span
}

// ====== Helpers (lookahead, bump, expectations) ======

func (p *Parser) at(k lexers.TokenKind) bool {
	return p.la != nil && p.la.Kind == k
}

func (p *Parser) atAny(ks ...lexers.TokenKind) bool {
	if p.la == nil {
		return false
	}
	for _, k := range ks {
		if p.la.Kind == k {
			return true
		}
	}
	return false
}

func (p *Parser) bump() *lexers.Token {
	tok := p.la
	if tok != nil {
		p.last = convertSpan(tok.Span)
	}
	p.la = p.lx.Next()
	return tok
}

func (p *Parser) want(k lexers.TokenKind) *TokenNode {
	if p.at(k) {
		return &TokenNode{Tok: p.bump()}
	}
	p.errorExpected(k, p.la)
	return p.synthToken(k)
}

func (p *Parser) wantOneOf(kinds ...lexers.TokenKind) *TokenNode {
	for _, k := range kinds {
		if p.at(k) {
			return &TokenNode{Tok: p.bump()}
		}
	}
	p.errorExpectedSet(kinds, p.la)
	// Synthesize the first expected kind as representative.
	return p.synthToken(kinds[0])
}

func (p *Parser) synthToken(k lexers.TokenKind) *TokenNode {
	ins := p.insertionSpan(p.la)
	tok := &lexers.Token{
		Kind: k,
		Span: convertBack(ins),
		// TODO: consider attaching synthetic trivia markers if needed by tools.
	}
	return &TokenNode{Tok: tok}
}

func (p *Parser) errorExpected(k lexers.TokenKind, found *lexers.Token) {
	var fdesc string
	if found == nil {
		fdesc = "EOF"
	} else {
		fdesc = found.Kind.String()
	}
	span := p.insertionSpan(found)
	p.addError(span, fmt.Sprintf("expected %s, found %s", k.String(), fdesc))
}

func (p *Parser) errorExpectedSet(ks []lexers.TokenKind, found *lexers.Token) {
	var fdesc string
	if found == nil {
		fdesc = "EOF"
	} else {
		fdesc = found.Kind.String()
	}
	span := p.insertionSpan(found)
	p.addError(span, fmt.Sprintf("expected one of %v, found %s", kindsToStrings(ks), fdesc))
}

func kindsToStrings(ks []lexers.TokenKind) []string {
	out := make([]string, 0, len(ks))
	for _, k := range ks {
		out = append(out, k.String())
	}
	return out
}

// recoverTo skips tokens until one of the sync kinds is seen or EOF.
func (p *Parser) recoverTo(sync ...lexers.TokenKind) {
	for p.la != nil && !p.atAny(sync...) {
		p.bump()
	}
}

func (p *Parser) synthesizeEOF() *TokenNode {
	// Create a zero-width EOF token node at end-of-input.
	// NOTE: lexers.TokenKind may not have an EOF kind. If it does, prefer that.
	// If not, choose a sentinel kind we've agreed on in the lexers package.
	// TODO: replace lexers.TokenKindEOF with the actual EOF kind name (if any).
	ins := p.insertionSpan(nil)
	tok := &lexers.Token{
		Kind: lexers.EOF, // TODO: confirm the name/availability in lexers.
		Span: convertBack(ins),
	}
	return &TokenNode{Tok: tok}
}

// ====== Productions ======

// parseHeader parses: Current Turn <Month> <Year> ( # <Number> )
func (p *Parser) parseHeader() *Header {
	h := &Header{}
	h.KwCurrent = p.want(lexers.KeywordCurrent)
	h.KwTurn = p.want(lexers.KeywordTurn)
	h.Month = p.wantOneOf(lexers.MonthName, lexers.Identifier)
	h.Year = p.want(lexers.Number)
	h.LParen = p.want(lexers.LParen)
	h.Hash = p.want(lexers.Hash)
	// Turn number is strictly digits in CST; see spec.
	turn := p.want(lexers.Number)
	// OPTIONAL one-token recovery: if a non-number slipped here, consume it and note.
	if turn.Tok.Kind != lexers.Number && p.la != nil && p.la.Kind == lexers.Number {
		// We synthesized a number; consume the bad token to avoid follow-on errors.
		p.addError(turn.Span(), "invalid turn number token; digits only")
		_ = p.bump()
	}
	h.TurnNo = turn
	h.RParen = p.want(lexers.RParen)
	h.span = coverNodesAndTokens(h.KwCurrent, h.RParen)
	return h
}

// parseBadTopLevel captures unknown top-level tokens until a sync point.
func (p *Parser) parseBadTopLevel() *BadTopLevel {
	start := p.la
	b := &BadTopLevel{}
	// Sync points for top-level: next header start or EOF.
	for p.la != nil && !p.at(lexers.KeywordCurrent) {
		b.Tokens = append(b.Tokens, &TokenNode{Tok: p.bump()})
	}
	if len(b.Tokens) > 0 {
		b.span = Span{
			Start: convertSpan(b.Tokens[0].Tok.Span).Start,
			End:   convertSpan(b.Tokens[len(b.Tokens)-1].Tok.Span).End,
			// Line/Col values here are approximate; callers rarely consume them
			// from Bad nodes. TODO: compute precise min/max line/col if needed.
		}
	} else {
		// Shouldn't happen (we only call with la != nil), but keep safe.
		b.span = p.insertionSpan(start)
	}
	p.addError(b.span, "unrecognized top-level construct; recovering")
	return b
}

// ====== Span covering ======

// coverNodesAndTokens computes a Span covering the provided nodes/tokens.
// Accepts nils; returns zero Span if nothing is provided.
func coverNodesAndTokens(parts ...interface{}) Span {
	var (
		ok    bool
		first = true
		out   Span
	)
	for _, part := range parts {
		if part == nil {
			continue
		}
		var s Span
		switch v := part.(type) {
		case *TokenNode:
			if v == nil || v.Tok == nil {
				continue
			}
			s = convertSpan(v.Tok.Span)
			ok = true
		case Node:
			if v == nil {
				continue
			}
			s = v.Span()
			ok = true
		}
		if !ok {
			continue
		}
		if first {
			out = s
			first = false
		} else {
			if s.Start < out.Start {
				out.Start = s.Start
			}
			if s.End > out.End {
				out.End = s.End
			}
			// NOTE: Line/Col are left as-is; for multi-line covers you may want
			// to carry the start line/col from the first and compute end line/col.
			// TODO: enhance if downstream tools rely on outer Line/Col.
		}
	}
	return out
}

// ====== TODOs / Integration Notes ======
//
// - Confirm lexers.Token fields and names (Kind, Span, LeadingTrivia, TrailingTrivia).
// - Confirm lexers.TokenKind names and presence of TokenKindEOF (or equivalent).
// - Decide whether newline is represented as trivia only or token; adjust sync.
// - Add per-production error budgets to prevent cascades (e.g., max 3 diags).
// - Add serialization utilities for CST (for golden tests).
// - Add pretty diagnostic formatting with source excerpts.
// - Implement additional productions beyond Header and wire them in top-level loop.
// - Consider exporting Header fields if external tools/tests need direct access.
