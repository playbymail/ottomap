// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package lexers_test

import (
	"testing"

	"github.com/playbymail/ottomap/internal/reports/lexers"
)

type tok struct {
	Kind string // use Kind.String() for comparison to avoid importing enum directly
	Text string
	// Minimal span checks (optional)
	Start, End int
}

type case_ struct {
	name  string
	input string
	want  []tok // expected significant tokens in order (ignore trivia here)
}

func TestLexer_Tokens_Simple(t *testing.T) {
	cases := []case_{
		{
			name:  "current_turn_happy",
			input: "Current Turn August 2025 (#123)",
			want: []tok{
				{Kind: "KeywordCurrent", Text: "Current"},
				{Kind: "KeywordTurn", Text: "Turn"},
				{Kind: "MonthName", Text: "August"},
				{Kind: "Number", Text: "2025"},
				{Kind: "LParen", Text: "("},
				{Kind: "Hash", Text: "#"},
				{Kind: "Number", Text: "123"},
				{Kind: "RParen", Text: ")"},
			},
		},
		{
			name:  "bad_month_and_hash",
			input: "Current Turn Agust 2025 (##12a3",
			want: []tok{
				{Kind: "KeywordCurrent", Text: "Current"},
				{Kind: "KeywordTurn", Text: "Turn"},
				{Kind: "Identifier", Text: "Agust"},
				{Kind: "Number", Text: "2025"},
				{Kind: "LParen", Text: "("},
				{Kind: "Hash", Text: "#"},
				{Kind: "Hash", Text: "#"},
				{Kind: "Identifier", Text: "12a3"},
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
