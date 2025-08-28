// Copyright (c) 2025 Michael D Henderson. All rights reserved.

//go:build test || !release

// Package asttest provides helpers for AST golden snapshot tests.
// Keep this lightweight and test-focused. Not for production use.
//
// Purpose: stable JSON snapshot for AST + diagnostics
// Handles: File, Header, and AST diagnostics
package asttest

import (
	"encoding/json"
	"sort"

	"github.com/playbymail/ottomap/internal/parsers/ast"
)

type span struct {
	Start int `json:"start"`
	End   int `json:"end"`
	Line  int `json:"line"`
	Col   int `json:"col"`
}

type header struct {
	Node         string `json:"node"`
	Span         span   `json:"span"`
	Month        int    `json:"month"`
	Year         int    `json:"year"`
	Turn         int    `json:"turn"`
	RawMonth     string `json:"rawMonth"`
	RawYear      string `json:"rawYear"`
	RawTurn      string `json:"rawTurn"`
	UnknownMonth bool   `json:"unknownMonth"`
	InvalidYear  bool   `json:"invalidYear"`
	InvalidTurn  bool   `json:"invalidTurn"`
}

type file struct {
	Node   string   `json:"node"`
	Span   span     `json:"span"`
	Header []header `json:"headers"`
}

type diag struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	Span     span   `json:"span"`
}

type snapshot struct {
	AST   file   `json:"ast"`
	Diags []diag `json:"diagnostics"`
}

func Snapshot(af *ast.File, diags []ast.Diagnostic) ([]byte, error) {
	s := snapshot{
		AST: file{
			Node: "File",
			Span: toSpan(af.NodeSpan()),
		},
	}
	for _, h := range af.Headers {
		s.AST.Header = append(s.AST.Header, header{
			Node:         "Header",
			Span:         toSpan(h.NodeSpan()),
			Month:        h.Month,
			Year:         h.Year,
			Turn:         h.Turn,
			RawMonth:     h.RawMonth,
			RawYear:      h.RawYear,
			RawTurn:      h.RawTurn,
			UnknownMonth: h.UnknownMonth,
			InvalidYear:  h.InvalidYear,
			InvalidTurn:  h.InvalidTurn,
		})
	}

	// stable diag order: code then span.start
	sort.Slice(diags, func(i, j int) bool {
		if diags[i].Code != diags[j].Code {
			return string(diags[i].Code) < string(diags[j].Code)
		}
		return diags[i].Span.Start < diags[j].Span.Start
	})
	for _, d := range diags {
		s.Diags = append(s.Diags, diag{
			Severity: sevName(d.Severity),
			Code:     string(d.Code),
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

func toSpan(sp ast.Span) span { return span{sp.Start, sp.End, sp.Line, sp.Col} }

func sevName(s ast.Severity) string {
	switch s {
	case ast.SeverityError:
		return "error"
	case ast.SeverityWarning:
		return "warning"
	case ast.SeverityInfo:
		return "info"
	default:
		return "unknown"
	}
}
