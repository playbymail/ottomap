// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package lexers_test

import (
	"testing"

	"github.com/playbymail/ottomap/internal/parsers/lexers"
)

type tok struct {
	Kind string // use Kind.String() for comparison to avoid importing enum directly
	Text string
	// Minimal span checks (optional)
	Start, End int
}

type testcase struct {
	name  string
	input string
	want  []tok // expected significant tokens in order (ignore trivia here)
}

func TestLexer_SignificantTokenStreams(t *testing.T) {
	cases := []testcase{
		{
			name:  "tribe_header_happy",
			input: "Tribe 0987, , Current Hex = OO 0202, (Previous Hex = OO 0202)",
			want: []tok{
				{Kind: "Identifier", Text: "Tribe"},
				{Kind: "Number", Text: "0987"},
				{Kind: "Comma", Text: ","},
				{Kind: "Comma", Text: ","},
				{Kind: "KeywordCurrent", Text: "Current"},
				{Kind: "Identifier", Text: "Hex"},
				{Kind: "Equal", Text: "="},
				{Kind: "Identifier", Text: "OO"},
				{Kind: "Number", Text: "0202"},
				{Kind: "Comma", Text: ","},
				{Kind: "LParen", Text: "("},
				{Kind: "Identifier", Text: "Previous"},
				{Kind: "Identifier", Text: "Hex"},
				{Kind: "Equal", Text: "="},
				{Kind: "Identifier", Text: "OO"},
				{Kind: "Number", Text: "0202"},
				{Kind: "RParen", Text: ")"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lx := lexers.New(1, 1, []byte(tc.input))
			var got []tok
			for {
				tk := lx.Next()
				if tk == nil {
					break
				}
				// Skip trivia: we’re validating significant token stream first.
				k := tk.Kind.String()
				txt := string([]byte(tc.input)[tk.Span.Start:tk.Span.End])
				got = append(got, tok{Kind: k, Text: txt, Start: tk.Span.Start, End: tk.Span.End})
			}
			if len(got) != len(tc.want) {
				t.Fatalf("len(tokens)=%d, want %d\n got=%v", len(got), len(tc.want), got)
			}
			for i := range tc.want {
				if got[i].Kind != tc.want[i].Kind || got[i].Text != tc.want[i].Text {
					t.Fatalf("tok[%d]=(%s,%q), want (%s,%q)", i, got[i].Kind, got[i].Text, tc.want[i].Kind, tc.want[i].Text)
				}
			}
		})
	}
}
