// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package ast builds a normalized AST from the lossless CST produced by the cst package.
package ast

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/playbymail/ottomap/internal/parsers/cst"
)

// ====== Public API ======

// Parse parses the input into a CST and then transforms it into an AST.
// It never panics. Returns an AST and AST-level diagnostics.
// NOTE: By default we do NOT merge CST diagnostics; see TODO below if you want to.
func Parse(input []byte) (*File, []Diagnostic) {
	cfile, _ := cst.ParseFile(input)
	afile, adiags := FromCST(cfile, input)
	return afile, adiags
}

// FromCST transforms a CST file into an AST file (without invoking the CST parser).
func FromCST(cf *cst.File, input []byte) (*File, []Diagnostic) {
	b := &builder{input: input}
	return b.file(cf)
}

// ====== Diagnostics (AST-level) ======

type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
	SeverityInfo
)

type Code string

const (
	CodeMonthUnknown   Code = "E_MONTH"
	CodeYearInvalid    Code = "E_YEAR"
	CodeYearOutOfRange      = "E_YEAR_RANGE"
	CodeTurnInvalid    Code = "E_TURN"

	// Optional warning codes (policy-dependent):

	CodeMissingHash Code = "W_MISSING_HASH"
	CodeDupHeader   Code = "W_DUPLICATE_HEADER"
)

// Span is copied from CST spans (not aliased) for stability.
type Span struct {
	Start int
	End   int
	Line  int
	Col   int
}

type Diagnostic struct {
	Severity Severity
	Code     Code
	Span     Span
	Message  string
	Notes    []string
}

// ====== AST Node Interfaces & Kinds ======

type Kind int

const (
	KindFile Kind = iota
	KindHeader
)

type Node interface {
	NodeKind() Kind
	NodeSpan() Span
	Source() Source
}

// Source ties an AST node back to its CST origin (opaque to callers).
type Source struct {
	FileSpan Span
	Origin   any // typically *cst.Header, *cst.File, etc.
}

// ====== AST Root ======

type File struct {
	Headers []*Header
	src     Source
}

func (f *File) NodeKind() Kind { return KindFile }
func (f *File) NodeSpan() Span { return f.src.FileSpan }
func (f *File) Source() Source { return f.src }

// ====== AST Header ======

type Header struct {
	Month int // 1..12 (0 if unknown)
	Year  int
	Turn  int

	RawMonth string
	RawYear  string
	RawTurn  string

	UnknownMonth bool
	InvalidYear  bool
	InvalidTurn  bool

	src Source
}

func (h *Header) NodeKind() Kind { return KindHeader }
func (h *Header) NodeSpan() Span { return h.src.FileSpan }
func (h *Header) Source() Source { return h.src }

// ====== Builder ======

type builder struct {
	input []byte
	diags []Diagnostic
}

func (b *builder) file(cf *cst.File) (*File, []Diagnostic) {
	out := &File{}

	// Transform recognizable top-level CST nodes.
	for _, d := range cf.Decls {
		switch n := d.(type) {
		case *cst.Header:
			out.Headers = append(out.Headers, b.header(n))
		default:
			// TODO: handle additional CST nodes (sections) as they are introduced.
		}
	}

	// Compute file span from children (if any).
	if len(out.Headers) > 0 {
		out.src.FileSpan = out.Headers[0].NodeSpan()
		for _, h := range out.Headers[1:] {
			out.src.FileSpan = cover(out.src.FileSpan, h.NodeSpan())
		}
	} else {
		// Fallback to CST root span if available (optional).
		out.src.FileSpan = toSpan(cf.Span())
		out.src.Origin = cf
	}

	// TODO (policy): if multiple headers should be constrained, emit warnings:
	// if len(out.Headers) > 1 {
	//     b.warn(out.Headers[1].NodeSpan(), CodeDupHeader, "duplicate header; later headers may be ignored")
	// }

	return out, b.diags
}

func (b *builder) header(ch *cst.Header) *Header {
	h := &Header{
		src: Source{
			FileSpan: toSpan(ch.Span()),
			Origin:   ch,
		},
	}

	// ---- Month ----
	rawMonth := tokenText(b.input, ch.Month)
	h.RawMonth = rawMonth
	if m, ok := normalizeMonth(rawMonth); ok {
		h.Month = m
	} else {
		h.UnknownMonth = true
		h.Month = 0
		b.err(tokenSpan(ch.Month), CodeMonthUnknown, "unknown month %q", rawMonth)
	}

	// ---- Year ----
	rawYear := tokenText(b.input, ch.Year)
	h.RawYear = rawYear
	if y, ok := parseYear(rawYear); ok {
		if !validYearRange(y) {
			h.InvalidYear = true
			h.Year = y
			b.err(tokenSpan(ch.Year), CodeYearOutOfRange, "year %d out of range", y)
		} else {
			h.Year = y
		}
	} else {
		h.InvalidYear = true
		b.err(tokenSpan(ch.Year), CodeYearInvalid, "invalid year %q", rawYear)
	}

	// ---- Turn ----
	rawTurn := tokenText(b.input, ch.TurnNo)
	h.RawTurn = rawTurn
	if t, ok := parsePositiveInt(rawTurn); ok && t >= 1 {
		h.Turn = t
	} else {
		h.InvalidTurn = true
		b.err(tokenSpan(ch.TurnNo), CodeTurnInvalid, "invalid turn number %q", rawTurn)
	}

	// ---- Optional policy-level warning about missing '#' (synthesized in CST) ----
	// If you choose to surface this at AST-level:
	// if isSynthetic(ch.Hash) {
	//     b.warn(tokenSpan(ch.Hash), CodeMissingHash, "missing '#' before turn number")
	// }

	return h
}

// ====== Builder diagnostics ======

func (b *builder) err(sp Span, code Code, format string, args ...any) {
	b.diags = append(b.diags, Diagnostic{
		Severity: SeverityError,
		Code:     code,
		Span:     sp,
		Message:  fmt.Sprintf(format, args...),
	})
}

func (b *builder) warn(sp Span, code Code, format string, args ...any) {
	b.diags = append(b.diags, Diagnostic{
		Severity: SeverityWarning,
		Code:     code,
		Span:     sp,
		Message:  fmt.Sprintf(format, args...),
	})
}

// ====== Helpers: CST interop ======

func toSpan(s cst.Span) Span {
	return Span{Start: s.Start, End: s.End, Line: s.Line, Col: s.Col}
}

func tokenSpan(tn *cst.TokenNode) Span {
	if tn == nil || tn.Tok == nil {
		return Span{} // unknown; callers should tolerate zero span
	}
	return toSpan(cst.Span(tn.Tok.Span))
}

func tokenText(input []byte, tn *cst.TokenNode) string {
	if tn == nil || tn.Tok == nil {
		return ""
	}
	s := tn.Tok.Span
	if s.Start < 0 || s.End < s.Start || s.End > len(input) {
		return "" // defensive; malformed span
	}
	return string(input[s.Start:s.End])
}

// isSynthetic returns true if the token's span is zero-width (CST synthesis).
func isSynthetic(tn *cst.TokenNode) bool {
	if tn == nil || tn.Tok == nil {
		return true
	}
	s := tn.Tok.Span
	return s.Start == s.End
}

// cover returns the minimal span that covers a and b.
func cover(a, b Span) Span {
	if a.Start == 0 && a.End == 0 {
		return b
	}
	if b.Start == 0 && b.End == 0 {
		return a
	}
	out := a
	if b.Start < out.Start {
		out.Start = b.Start
		out.Line = b.Line
		out.Col = b.Col
	}
	if b.End > out.End {
		out.End = b.End
	}
	// NOTE: We keep Line/Col from the earliest start; end line/col are not tracked.
	return out
}

// ====== Normalization & Parsing Utilities ======

var months = map[string]int{
	"january": 1, "jan": 1,
	"february": 2, "feb": 2,
	"march": 3, "mar": 3,
	"april": 4, "apr": 4,
	"may":  5,
	"june": 6, "jun": 6,
	"july": 7, "jul": 7,
	"august": 8, "aug": 8,
	"september": 9, "sep": 9, "sept": 9,
	"october": 10, "oct": 10,
	"november": 11, "nov": 11,
	"december": 12, "dec": 12,
}

func normalizeMonth(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	key := strings.ToLower(strings.TrimSpace(s))
	m, ok := months[key]
	return m, ok
}

func parseYear(s string) (int, bool) {
	// Accept only base-10 integers with optional surrounding whitespace.
	v, err := strconv.Atoi(strings.TrimSpace(s))
	return v, err == nil
}

func validYearRange(y int) bool {
	// TODO: adjust policy range as needed for your domain.
	return y >= 1900 && y <= 2200
}

func parsePositiveInt(s string) (int, bool) {
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, false
	}
	return v, true
}

// ====== TODOs / Integration Notes ======
//
// - Decide whether to merge CST diagnostics with AST diagnostics in Parse():
//     cfile, cdiags := cst.ParseFile(input)
//     afile, adiags := FromCST(cfile, input)
//     return afile, append(convertCSTDiags(cdiags), adiags...)
//   Implement convertCSTDiags if you want a unified stream.
//
// - Define/implement policy for multiple headers (keep all vs. first/last with warnings).
//
// - Add additional AST node types and their CST → AST transformations as new sections are added.
//
// - Consider exposing JSON/YAML serialization of the AST for testing or tooling.
//
// - Consider attaching stable IDs to AST nodes for cross-references in later phases.
//
// - If you need i18n month names/abbreviations, extend normalizeMonth accordingly.
//
// - If CST tokens can be locale-specific or case-variant, add normalization hooks here.
//
// - Add golden tests that assert: (1) AST shape and values, (2) diagnostics (codes/messages/spans).
