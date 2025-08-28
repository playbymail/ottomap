// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package lexers_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/playbymail/ottomap/internal/reports/lexers"
)

var update = flag.Bool("update", false, "update lexer golden snapshots")

type snapToken struct {
	Kind     string       `json:"kind"`
	Text     string       `json:"text"`
	Span     span         `json:"span"`
	Leading  []triviaSnap `json:"leadingTrivia,omitempty"`
	Trailing []triviaSnap `json:"trailingTrivia,omitempty"`
}

type triviaSnap struct {
	Kind string `json:"kind"`
	Text string `json:"text"`
	Span span   `json:"span"`
}

type span struct{ Start, End, Line, Col int }
type snapshot struct {
	Tokens []snapToken `json:"tokens"`
}

func TestLexer_Golden(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"happy_current_turn", "Current Turn August 2025 (#123)\n"},
		{"bad_month_hash_number", "Current Turn Agust 2025 (##12a3\n"},
		{"extra_commas_comment", "Current,,  Turn /*c*/ August 2025 (#7))\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lx := lexers.New(1, 1, []byte(tc.input))
			var snap snapshot
			for {
				tk := lx.Next()
				if tk == nil {
					break
				}
				s := snapToken{
					Kind: tk.Kind.String(),
					Text: tc.input[tk.Span.Start:tk.Span.End],
					Span: span{tk.Span.Start, tk.Span.End, tk.Span.Line, tk.Span.Col},
				}
				for _, tr := range tk.LeadingTrivia {
					s.Leading = append(s.Leading, triviaSnap{
						Kind: tr.Kind.String(),
						Text: tc.input[tr.Span.Start:tr.Span.End],
						Span: span{tr.Span.Start, tr.Span.End, tr.Span.Line, tr.Span.Col},
					})
				}
				for _, tr := range tk.TrailingTrivia {
					s.Trailing = append(s.Trailing, triviaSnap{
						Kind: tr.Kind.String(),
						Text: tc.input[tr.Span.Start:tr.Span.End],
						Span: span{tr.Span.Start, tr.Span.End, tr.Span.Line, tr.Span.Col},
					})
				}
				snap.Tokens = append(snap.Tokens, s)
			}
			got := mustJSON(t, snap)
			path := filepath.Join("internal", "reports", "lexers", "testdata", sanitize(tc.name)+".lexer.json")
			if *update {
				mustWrite(t, path, got)
			}
			want := mustRead(t, path)
			if !bytes.Equal(bytes.TrimSpace(got), bytes.TrimSpace(want)) {
				t.Fatalf("lexer snapshot mismatch\nGOT:\n%s\nWANT:\n%s", got, want)
			}
		})
	}
}

// --- Utilities --------------------------------------------------------------

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("json: %v", err)
	}
	return append(b, '\n')
}

func mustWrite(t *testing.T, p string, b []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustRead(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v (run with -update)", p, err)
	}
	return b
}

func sanitize(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "_")
	return s
}
