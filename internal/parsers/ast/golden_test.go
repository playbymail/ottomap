// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package ast_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/playbymail/ottomap/internal/parsers/ast"
	"github.com/playbymail/ottomap/internal/parsers/ast/asttest"
)

// go test ./ast -run TestAST_Golden -update
var update = flag.Bool("update", false, "update golden files")

// --- Public test entry -------------------------------------------------------

func TestAST_Golden(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
	}{
		{
			name:  "happy_current_turn",
			input: "Current Turn August 2025 (#123)\n",
		},
		{
			name:  "unknown_month_and_bad_turn",
			input: "Current Turn Agust 2025 (#12a3)\n",
		},
		{
			name:  "year_out_of_range",
			input: "Current Turn July 3025 (#5)\n",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			file, diags := ast.Parse([]byte(tc.input))

			// create a snapshot
			snap, err := asttest.Snapshot(file, diags)
			if err != nil {
				t.Fatal(err)
			}
			gotJSON := mustJSON(t, snap)

			goldenPath := filepath.Join("internal", "reports", "ast", "testdata", sanitize(tc.name)+".golden.json")
			if *update {
				mustWriteFile(t, goldenPath, gotJSON)
			}

			// compare snapshot to golden
			wantJSON := mustReadFile(t, goldenPath)
			// Byte compare for strictness; failures show a unified diff in most editors/CI.
			if !bytes.Equal(bytes.TrimSpace(gotJSON), bytes.TrimSpace(wantJSON)) {
				t.Fatalf("AST snapshot mismatch for %s\nGOT:\n%s\n\nWANT:\n%s",
					tc.name, gotJSON, wantJSON)
			}
		})
	}
}

// --- Utilities --------------------------------------------------------------

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	buf, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("json marshal: %v", err)
	}
	return append(buf, '\n') // newline at EOF for nicer diffs
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v\nTip: run with -update to create golden.", path, err)
	}
	return b
}

func mustWriteFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func sanitize(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	return s
}
