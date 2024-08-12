// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package coords_test

import (
	"github.com/playbymail/ottomap/internal/coords"
	"testing"
)

func TestStringToGridCoordinates(t *testing.T) {
	tests := []struct {
		id         int
		input      string
		wantResult coords.Grid
		wantErr    bool
		wantString string
	}{
		// Add as many test cases as you want
		{1001, "AA 0101", coords.Grid{BigMapRow: 0, BigMapColumn: 0, GridColumn: 1, GridRow: 1}, false, "AA 0101"},
		{1002, "AZ 3001", coords.Grid{BigMapRow: 0, BigMapColumn: 25, GridColumn: 30, GridRow: 1}, false, "AZ 3001"},
		{1003, "ZA 0121", coords.Grid{BigMapRow: 25, BigMapColumn: 0, GridColumn: 1, GridRow: 21}, false, "ZA 0121"},
		{1004, "ZZ 3021", coords.Grid{BigMapRow: 25, BigMapColumn: 25, GridColumn: 30, GridRow: 21}, false, "ZZ 3021"},
		{2001, "## 0101", coords.Grid{BigMapRow: 0, BigMapColumn: 0, GridColumn: 1, GridRow: 1}, false, "AA 0101"},
		{2002, "## 3001", coords.Grid{BigMapRow: 0, BigMapColumn: 0, GridColumn: 30, GridRow: 1}, false, "AA 3001"},
		{2003, "## 0121", coords.Grid{BigMapRow: 0, BigMapColumn: 0, GridColumn: 1, GridRow: 21}, false, "AA 0121"},
		{2004, "## 3021", coords.Grid{BigMapRow: 0, BigMapColumn: 0, GridColumn: 30, GridRow: 21}, false, "AA 3021"},
		{2005, "## 1206", coords.Grid{BigMapRow: 0, BigMapColumn: 0, GridColumn: 12, GridRow: 6}, false, "AA 1206"},
		{2006, "## 1306", coords.Grid{BigMapRow: 0, BigMapColumn: 0, GridColumn: 13, GridRow: 6}, false, "AA 1306"},
		{3101, "AA0101", coords.Grid{}, true, "AA 0000"},
		{3102, "AA-0101", coords.Grid{}, true, "AA 0000"},
		{3201, "aB 1230", coords.Grid{}, true, "AA 0000"},
		{3301, "Ab 1230", coords.Grid{}, true, "AA 0000"},
		{3401, "AA 0001", coords.Grid{}, true, "AA 0000"},
		{3402, "AA 3100", coords.Grid{}, true, "AA 0000"},
		{3501, "AA 0100", coords.Grid{}, true, "AA 0000"},
		{3502, "AA 0122", coords.Grid{}, true, "AA 0000"},
	}

	for _, tc := range tests {
		gotResult, gotErr := coords.StringToGridCoords(tc.input)
		if tc.wantErr == true {
			if gotErr == nil {
				t.Errorf("%d: errCheck: got nil, want !nil", tc.id)
			}
			continue
		}
		if gotErr != nil {
			t.Errorf("%d: errCheck: got %v, want nil", tc.id, gotErr)
			continue
		}
		checkString := true
		if gotResult.BigMapRow != tc.wantResult.BigMapRow {
			checkString = false
			t.Errorf("%d: %q: bigMapRow   : got %6d, want %6d", tc.id, tc.input, gotResult.BigMapRow, tc.wantResult.BigMapRow)
		}
		if gotResult.BigMapColumn != tc.wantResult.BigMapColumn {
			checkString = false
			t.Errorf("%d: %q: bigMapColumn: got %6d, want %6d", tc.id, tc.input, gotResult.BigMapColumn, tc.wantResult.BigMapColumn)
		}
		if gotResult.GridColumn != tc.wantResult.GridColumn {
			checkString = false
			t.Errorf("%d: %q: gridColumn  : got %6d, want %6d", tc.id, tc.input, gotResult.GridColumn, tc.wantResult.GridColumn)
		}
		if gotResult.GridRow != tc.wantResult.GridRow {
			checkString = false
			t.Errorf("%d: %q: gridRow     : got %6d, want %6d", tc.id, tc.input, gotResult.GridRow, tc.wantResult.GridRow)
		}
		if checkString && gotResult.String() != tc.wantString {
			t.Errorf("%d: %q: got %q, want %q", tc.id, tc.input, gotResult.String(), tc.wantString)
		}
	}
}

func TestGridToMapCoordinates(t *testing.T) {
	tests := []struct {
		id         int
		input      string
		wantResult coords.Map
		wantString string
	}{
		{1002, "AA 0101", coords.Map{Column: 0, Row: 0}, "(0, 0)"},
		{1003, "AA 3001", coords.Map{Column: 29, Row: 0}, "(29, 0)"},
		{1004, "AA 0121", coords.Map{Column: 0, Row: 20}, "(0, 20)"},
		{1005, "AA 3021", coords.Map{Column: 29, Row: 20}, "(29, 20)"},
		{1006, "AB 0101", coords.Map{Column: 30, Row: 0}, "(30, 0)"},
		{1007, "AB 3001", coords.Map{Column: 59, Row: 0}, "(59, 0)"},
		{1008, "AB 0121", coords.Map{Column: 30, Row: 20}, "(30, 20)"},
		{1009, "AB 3021", coords.Map{Column: 59, Row: 20}, "(59, 20)"},
		{1010, "AZ 0101", coords.Map{Column: 750, Row: 0}, "(750, 0)"},
		{1011, "AZ 3001", coords.Map{Column: 779, Row: 0}, "(779, 0)"},
		{1012, "AZ 3021", coords.Map{Column: 779, Row: 20}, "(779, 20)"},
		{1013, "HP 1511", coords.Map{Column: 464, Row: 157}, "(464, 157)"},
		{1014, "ZA 0101", coords.Map{Column: 0, Row: 525}, "(0, 525)"},
		{1015, "ZA 3001", coords.Map{Column: 29, Row: 525}, "(29, 525)"},
		{1016, "ZA 0121", coords.Map{Column: 0, Row: 545}, "(0, 545)"},
		{1016, "ZA 3021", coords.Map{Column: 29, Row: 545}, "(29, 545)"},
		{1017, "ZZ 0101", coords.Map{Column: 750, Row: 525}, "(750, 525)"},
		{1017, "ZZ 3001", coords.Map{Column: 779, Row: 525}, "(779, 525)"},
		{1018, "ZZ 0121", coords.Map{Column: 750, Row: 545}, "(750, 545)"},
		{1018, "ZZ 3021", coords.Map{Column: 779, Row: 545}, "(779, 545)"},
	}

	for _, tc := range tests {
		gc, err := coords.StringToGridCoords(tc.input)
		if err != nil {
			t.Errorf("%d: errCheck: got %v, want nil", tc.id, err)
			continue
		}
		mc, err := gc.ToMapCoords()
		if err != nil {
			t.Errorf("%d: %q: %v", tc.id, tc.input, err)
			continue
		}
		if mc.Column != tc.wantResult.Column {
			t.Errorf("%d: %q: column  : got %6d, want %6d", tc.id, tc.input, mc.Column, tc.wantResult.Column)
		}
		if mc.Row != tc.wantResult.Row {
			t.Errorf("%d: %q: row     : got %6d, want %6d", tc.id, tc.input, mc.Row, tc.wantResult.Row)
		}
		if mc.String() != tc.wantString {
			t.Errorf("%d: %q: string  : got %q, want %q", tc.id, tc.input, mc.String(), tc.wantString)
		}
	}
}
