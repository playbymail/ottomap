// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package cst_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/playbymail/ottomap/internal/reports/cst"
	"github.com/playbymail/ottomap/internal/reports/cst/csttest"
)

var update = flag.Bool("update", false, "update CST golden snapshots")

type span struct{ Start, End, Line, Col int }

type tok struct {
	Kind string `json:"kind"`
	Text string `json:"text"`
	Span span   `json:"span"`
}

type headerSnap struct {
	KwCurrent tok  `json:"kwCurrent"`
	KwTurn    tok  `json:"kwTurn"`
	Month     tok  `json:"month"`
	Year      tok  `json:"year"`
	LParen    tok  `json:"lparen"`
	Hash      tok  `json:"hash"`
	TurnNo    tok  `json:"turnNo"`
	RParen    tok  `json:"rparen"`
	Span      span `json:"span"`
}

type diag struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Span     span   `json:"span"`
}

type snapshot struct {
	Nodes []any  `json:"nodes"` // mixed nodes (only Header in examples)
	Diags []diag `json:"diagnostics"`
}

func TestCST_Golden(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"header_ok", "Current Turn August 2025 (#123)\n"},
		{"header_missing_hash", "Current Turn August 2025 (123)\n"},
		{"header_bad_turn", "Current Turn August 2025 (#12a3)\n"},
		{"garbled_toplevel", "Currnet  Turn  Agust 2025 (##12a3\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.input)
			f, diags := cst.ParseFile(src)

			// create a snapshot
			snap, err := csttest.Snapshot(f, src, diags)
			if err != nil {
				t.Fatal(err)
			}
			gotJSON := mustJSON(t, snap)

			// load the golden
			path := filepath.Join("internal", "reports", "cst", "testdata", sanitize(tc.name)+".cst.json")
			if *update {
				mustWrite(t, path, gotJSON)
			}

			// compare snapshot to golden
			wantJSON := mustRead(t, path)
			// Byte compare for strictness; failures show a unified diff in most editors/CI.
			if !bytes.Equal(bytes.TrimSpace(gotJSON), bytes.TrimSpace(wantJSON)) {
				t.Fatalf("CST snapshot mismatch\nGOT:\n%s\nWANT:\n%s", gotJSON, wantJSON)
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
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
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
