// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package ast_test

import (
	"testing"

	"github.com/playbymail/ottomap/internal/reports/ast"
)

// How to use it
//
// * Add new cases quickly: `"Current Turn Feb 1899 (#1)"` to test `E_YEAR_RANGE`, `"Current Turn  202X (#42)"` for `E_YEAR`, etc.
// * Each case just fills in the **AST.Header you expect** and the **diagnostic codes**.
// * This pattern lets you scale out tests easily while keeping assertions explicit.

type headerCase struct {
	name   string
	input  string
	expect ast.Header // expected normalized fields + flags
	diags  []ast.Code // expected diagnostic codes
}

func TestHeaderCrosswalk(t *testing.T) {
	cases := []headerCase{
		{
			name:  "happy path",
			input: "Current Turn August 2025 (#123)",
			expect: ast.Header{
				Month: 8, Year: 2025, Turn: 123,
				RawMonth: "August", RawYear: "2025", RawTurn: "123",
			},
			diags: nil,
		},
		{
			name:  "bad month + bad turn",
			input: "Current Turn Agust 2025 (#12a3)",
			expect: ast.Header{
				Month: 0, Year: 2025, Turn: 0,
				RawMonth: "Agust", RawYear: "2025", RawTurn: "12a3",
				UnknownMonth: true, InvalidTurn: true,
			},
			diags: []ast.Code{ast.CodeMonthUnknown, ast.CodeTurnInvalid},
		},
		{
			name:  "year out of range",
			input: "Current Turn July 3025 (#5)",
			expect: ast.Header{
				Month: 7, Year: 3025, Turn: 5,
				RawMonth: "July", RawYear: "3025", RawTurn: "5",
				InvalidYear: true,
			},
			diags: []ast.Code{ast.CodeYearOutOfRange},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			file, diags := ast.Parse([]byte(tc.input))
			if len(file.Headers) == 0 {
				t.Fatalf("no headers parsed")
			}
			got := file.Headers[0]

			// Check normalized fields
			if got.Month != tc.expect.Month {
				t.Errorf("Month = %d, want %d", got.Month, tc.expect.Month)
			}
			if got.Year != tc.expect.Year {
				t.Errorf("Year = %d, want %d", got.Year, tc.expect.Year)
			}
			if got.Turn != tc.expect.Turn {
				t.Errorf("Turn = %d, want %d", got.Turn, tc.expect.Turn)
			}
			if got.RawMonth != tc.expect.RawMonth {
				t.Errorf("RawMonth = %q, want %q", got.RawMonth, tc.expect.RawMonth)
			}
			if got.RawYear != tc.expect.RawYear {
				t.Errorf("RawYear = %q, want %q", got.RawYear, tc.expect.RawYear)
			}
			if got.RawTurn != tc.expect.RawTurn {
				t.Errorf("RawTurn = %q, want %q", got.RawTurn, tc.expect.RawTurn)
			}
			if got.UnknownMonth != tc.expect.UnknownMonth {
				t.Errorf("UnknownMonth = %v, want %v", got.UnknownMonth, tc.expect.UnknownMonth)
			}
			if got.InvalidYear != tc.expect.InvalidYear {
				t.Errorf("InvalidYear = %v, want %v", got.InvalidYear, tc.expect.InvalidYear)
			}
			if got.InvalidTurn != tc.expect.InvalidTurn {
				t.Errorf("InvalidTurn = %v, want %v", got.InvalidTurn, tc.expect.InvalidTurn)
			}

			// Check diagnostic codes (order-agnostic simple contains check)
			for _, code := range tc.diags {
				found := false
				for _, d := range diags {
					if d.Code == code {
						found = true
					}
				}
				if !found {
					t.Errorf("expected diagnostic %q not found; got %+v", code, diags)
				}
			}
		})
	}
}
