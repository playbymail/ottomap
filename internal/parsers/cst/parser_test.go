// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package cst_test

import (
	"testing"

	"github.com/playbymail/ottomap/internal/parsers/cst"
)

func TestCST_HeaderHappy(t *testing.T) {
	src := []byte("Current Turn August 2025 (#123)\n")
	file, diags := cst.ParseFile(src)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if len(file.Decls) != 1 {
		t.Fatalf("Decls=%d, want 1", len(file.Decls))
	}
	h, ok := file.Decls[0].(*cst.Header)
	if !ok {
		t.Fatalf("first decl is %T, want *cst.Header", file.Decls[0])
	}
	if h.Month == nil || h.Month.Tok == nil || string(src[h.Month.Tok.Span.Start:h.Month.Tok.Span.End]) != "August" {
		t.Fatalf("bad month token")
	}
	if h.Hash == nil || h.Hash.Tok == nil {
		t.Fatalf("missing hash")
	}
}
