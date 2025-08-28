// Copyright (c) 2025 Michael D Henderson. All rights reserved.

//go:build test || !release

// Package csttest provides helpers for CST golden snapshot tests.
// Keep this lightweight and test-focused. Not for production use.
//
// Purpose: turn a *cst.File (+source bytes) into a compact JSON snapshot
// Handles: File, Header, BadTopLevel, TokenNode, and diagnostics
package csttest

import (
	"encoding/json"
	"sort"

	"github.com/playbymail/ottomap/internal/parsers/cst"
)

type span struct {
	Start int `json:"start"`
	End   int `json:"end"`
	Line  int `json:"line"`
	Col   int `json:"col"`
}

type tok struct {
	Kind string `json:"kind"`
	Text string `json:"text"`
	Span span   `json:"span"`
}

type headerSnap struct {
	Node string `json:"node"`
	Span span   `json:"span"`

	KwCurrent tok `json:"kwCurrent"`
	KwTurn    tok `json:"kwTurn"`
	Month     tok `json:"month"`
	Year      tok `json:"year"`
	LParen    tok `json:"lparen"`
	Hash      tok `json:"hash"`
	TurnNo    tok `json:"turnNo"`
	RParen    tok `json:"rparen"`
}

type badTopLevelSnap struct {
	Node   string `json:"node"`
	Span   span   `json:"span"`
	Tokens []tok  `json:"tokens"`
}

type diagSnap struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Span     span   `json:"span"`
}

type fileSnap struct {
	Nodes       []any      `json:"nodes"`
	EOF         tok        `json:"eof"`
	FileSpan    span       `json:"span"`
	Diagnostics []diagSnap `json:"diagnostics"`
}

// Snapshot marshals a CST file plus diags to pretty JSON for goldens.
func Snapshot(cf *cst.File, src []byte, diags []cst.Diagnostic) ([]byte, error) {
	s := fileSnap{
		FileSpan: toSpan(cf.Span()),
		EOF:      tokOf(src, cf.EOF),
	}
	for _, n := range cf.Decls {
		switch v := n.(type) {
		case *cst.Header:
			s.Nodes = append(s.Nodes, headerOf(src, v))
		case *cst.BadTopLevel:
			s.Nodes = append(s.Nodes, badTopOf(src, v))
		default:
			s.Nodes = append(s.Nodes, map[string]any{
				"node": n.Kind().String(),
				"span": toSpan(n.Span()),
			})
		}
	}

	// stable diags order: severity, message, span.start
	sort.Slice(diags, func(i, j int) bool {
		if diags[i].Severity != diags[j].Severity {
			return diags[i].Severity < diags[j].Severity
		}
		if diags[i].Message != diags[j].Message {
			return diags[i].Message < diags[j].Message
		}
		return diags[i].Span.Start < diags[j].Span.Start
	})
	for _, d := range diags {
		s.Diagnostics = append(s.Diagnostics, diagSnap{
			Severity: sevName(d.Severity),
			Message:  d.Message,
			Span:     toSpan(d.Span),
		})
	}

	out, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return nil, err
	}
	out = append(out, '\n')
	return out, nil
}

func headerOf(src []byte, h *cst.Header) headerSnap {
	return headerSnap{
		Node:      "Header",
		Span:      toSpan(h.Span()),
		KwCurrent: tokOf(src, h.KwCurrent),
		KwTurn:    tokOf(src, h.KwTurn),
		Month:     tokOf(src, h.Month),
		Year:      tokOf(src, h.Year),
		LParen:    tokOf(src, h.LParen),
		Hash:      tokOf(src, h.Hash),
		TurnNo:    tokOf(src, h.TurnNo),
		RParen:    tokOf(src, h.RParen),
	}
}

func badTopOf(src []byte, b *cst.BadTopLevel) badTopLevelSnap {
	out := badTopLevelSnap{
		Node: "BadTopLevel",
		Span: toSpan(b.Span()),
	}
	for _, t := range b.Tokens {
		out.Tokens = append(out.Tokens, tokOf(src, t))
	}
	return out
}

func tokOf(src []byte, tn *cst.TokenNode) tok {
	if tn == nil || tn.Tok == nil {
		return tok{Kind: "nil", Text: "", Span: span{}}
	}
	sp := tn.Tok.Span
	return tok{
		Kind: tn.Tok.Kind.String(),
		Text: safeSlice(src, sp.Start, sp.End),
		Span: span{Start: sp.Start, End: sp.End, Line: sp.Line, Col: sp.Col},
	}
}

func sevName(s cst.Severity) string {
	switch s {
	case cst.SeverityError:
		return "error"
	case cst.SeverityWarning:
		return "warning"
	case cst.SeverityInfo:
		return "info"
	default:
		return "unknown"
	}
}

func toSpan(s cst.Span) span { return span{s.Start, s.End, s.Line, s.Col} }

func safeSlice(b []byte, i, j int) string {
	if i < 0 || j < i || j > len(b) {
		return ""
	}
	return string(b[i:j])
}
