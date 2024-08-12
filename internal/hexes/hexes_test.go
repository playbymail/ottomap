// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package hexes_test

import (
	"github.com/playbymail/ottomap/internal/direction"
	"github.com/playbymail/ottomap/internal/hexes"
	"testing"
)

func TestHexToString(t *testing.T) {
	tests := []struct {
		name     string
		hex      hexes.Hex_t
		expected string
	}{
		{
			name:     "AA 0101",
			hex:      hexes.Hex_t{GridRow: 1, GridColumn: 1, Column: 1, Row: 1},
			expected: "AA 0101",
		},
		{
			name:     "AZ 3021",
			hex:      hexes.Hex_t{GridRow: 1, GridColumn: 1, Column: 1, Row: 1},
			expected: "AA 0101",
		},
		{
			name:     "ZA 0101",
			hex:      hexes.Hex_t{GridRow: 1, GridColumn: 1, Column: 1, Row: 1},
			expected: "AA 0101",
		},
		{
			name:     "ZZ 2101",
			hex:      hexes.Hex_t{GridRow: 1, GridColumn: 1, Column: 1, Row: 1},
			expected: "AA 0101",
		},
	}

	for _, test := range tests {
		result := test.hex.String()
		if result != test.expected {
			t.Errorf("Unexpected result. Got %q, expected %q", result, test.expected)
		}
	}
}

func TestHexAdd(t *testing.T) {
	tests := []struct {
		name     string
		hex      hexes.Hex_t
		dir      direction.Direction_e
		expected hexes.Hex_t
	}{
		// test moving from an even column hex
		{
			name:     "Move N from OO 1608",
			hex:      hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 16, Row: 8},
			dir:      direction.North,
			expected: hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 16, Row: 7},
		},
		{
			name:     "Move NE from OO 1608",
			hex:      hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 16, Row: 8},
			dir:      direction.NorthEast,
			expected: hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 17, Row: 8},
		},
		{
			name:     "Move SE from OO 1608",
			hex:      hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 16, Row: 8},
			dir:      direction.SouthEast,
			expected: hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 17, Row: 9},
		},
		{
			name:     "Move S from OO 1608",
			hex:      hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 16, Row: 8},
			dir:      direction.South,
			expected: hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 16, Row: 9},
		},
		{
			name:     "Move SW from OO 1608",
			hex:      hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 16, Row: 8},
			dir:      direction.SouthWest,
			expected: hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 15, Row: 9},
		},
		{
			name:     "Move NW from OO 1608",
			hex:      hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 16, Row: 8},
			dir:      direction.NorthWest,
			expected: hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 15, Row: 8},
		},
		// test moving from an odd column hex
		{
			name:     "Move N from OO 1708",
			hex:      hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 17, Row: 8},
			dir:      direction.North,
			expected: hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 17, Row: 7},
		},
		{
			name:     "Move NE from OO 1708",
			hex:      hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 17, Row: 8},
			dir:      direction.NorthEast,
			expected: hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 18, Row: 7},
		},
		{
			name:     "Move SE from OO 1708",
			hex:      hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 17, Row: 8},
			dir:      direction.SouthEast,
			expected: hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 18, Row: 8},
		},
		{
			name:     "Move S from OO 1708",
			hex:      hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 17, Row: 8},
			dir:      direction.South,
			expected: hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 17, Row: 9},
		},
		{
			name:     "Move SW from OO 1708",
			hex:      hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 17, Row: 8},
			dir:      direction.SouthWest,
			expected: hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 16, Row: 8},
		},
		{
			name:     "Move NW from OO 1708",
			hex:      hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 17, Row: 8},
			dir:      direction.NorthWest,
			expected: hexes.Hex_t{GridRow: 15, GridColumn: 15, Column: 16, Row: 7},
		},
		// Test moving from upper left corner of a grid
		{
			name:     "Move N from DT 0101",
			hex:      hexes.Hex_t{GridRow: 4, GridColumn: 20, Column: 1, Row: 1},
			dir:      direction.North,
			expected: hexes.Hex_t{GridRow: 3, GridColumn: 20, Column: 1, Row: 21},
		},
		{
			name:     "Move NE from DT 0101",
			hex:      hexes.Hex_t{GridRow: 4, GridColumn: 20, Column: 1, Row: 1},
			dir:      direction.NorthEast,
			expected: hexes.Hex_t{GridRow: 3, GridColumn: 20, Column: 2, Row: 21},
		},
		{
			name:     "Move SE from DT 0101",
			hex:      hexes.Hex_t{GridRow: 4, GridColumn: 20, Column: 1, Row: 1},
			dir:      direction.SouthEast,
			expected: hexes.Hex_t{GridRow: 4, GridColumn: 20, Column: 2, Row: 1},
		},
		{
			name:     "Move S from DT 0101",
			hex:      hexes.Hex_t{GridRow: 4, GridColumn: 20, Column: 1, Row: 1},
			dir:      direction.South,
			expected: hexes.Hex_t{GridRow: 4, GridColumn: 20, Column: 1, Row: 2},
		},
		{
			name:     "Move SW from DT 0101",
			hex:      hexes.Hex_t{GridRow: 4, GridColumn: 20, Column: 1, Row: 1},
			dir:      direction.SouthWest,
			expected: hexes.Hex_t{GridRow: 4, GridColumn: 19, Column: 30, Row: 1},
		},
		{
			name:     "Move NW from DT 0101",
			hex:      hexes.Hex_t{GridRow: 4, GridColumn: 20, Column: 1, Row: 1},
			dir:      direction.NorthWest,
			expected: hexes.Hex_t{GridRow: 3, GridColumn: 19, Column: 30, Row: 21},
		},
		// Test moving from upper right corner of a grid
		{
			name:     "Move N from DS 3001",
			hex:      hexes.Hex_t{GridRow: 4, GridColumn: 19, Column: 30, Row: 1},
			dir:      direction.North,
			expected: hexes.Hex_t{GridRow: 3, GridColumn: 19, Column: 30, Row: 21},
		},
		{
			name:     "Move NE from DS 3001",
			hex:      hexes.Hex_t{GridRow: 4, GridColumn: 19, Column: 30, Row: 1},
			dir:      direction.NorthEast,
			expected: hexes.Hex_t{GridRow: 4, GridColumn: 20, Column: 1, Row: 1},
		},
		{
			name:     "Move SE from DS 3001",
			hex:      hexes.Hex_t{GridRow: 4, GridColumn: 19, Column: 30, Row: 1},
			dir:      direction.SouthEast,
			expected: hexes.Hex_t{GridRow: 4, GridColumn: 20, Column: 1, Row: 2},
		},
		{
			name:     "Move S from DS 3001",
			hex:      hexes.Hex_t{GridRow: 4, GridColumn: 19, Column: 30, Row: 1},
			dir:      direction.South,
			expected: hexes.Hex_t{GridRow: 4, GridColumn: 19, Column: 30, Row: 2},
		},
		{
			name:     "Move SW from DS 3001",
			hex:      hexes.Hex_t{GridRow: 4, GridColumn: 19, Column: 30, Row: 1},
			dir:      direction.SouthWest,
			expected: hexes.Hex_t{GridRow: 4, GridColumn: 19, Column: 29, Row: 2},
		},
		{
			name:     "Move NW from DS 3001",
			hex:      hexes.Hex_t{GridRow: 4, GridColumn: 19, Column: 30, Row: 1},
			dir:      direction.NorthWest,
			expected: hexes.Hex_t{GridRow: 4, GridColumn: 19, Column: 29, Row: 1},
		},
		// Test moving from bottom right corner of a grid
		{
			name:     "Move N from CS 3021",
			hex:      hexes.Hex_t{GridRow: 3, GridColumn: 19, Column: 30, Row: 21},
			dir:      direction.North,
			expected: hexes.Hex_t{GridRow: 3, GridColumn: 19, Column: 30, Row: 20},
		},
		{
			name:     "Move NE from CS 3021",
			hex:      hexes.Hex_t{GridRow: 3, GridColumn: 19, Column: 30, Row: 21},
			dir:      direction.NorthEast,
			expected: hexes.Hex_t{GridRow: 3, GridColumn: 20, Column: 1, Row: 21},
		},
		{
			name:     "Move SE from CS 3021",
			hex:      hexes.Hex_t{GridRow: 3, GridColumn: 19, Column: 30, Row: 21},
			dir:      direction.SouthEast,
			expected: hexes.Hex_t{GridRow: 4, GridColumn: 20, Column: 1, Row: 1},
		},
		{
			name:     "Move S from CS 3021",
			hex:      hexes.Hex_t{GridRow: 3, GridColumn: 19, Column: 30, Row: 21},
			dir:      direction.South,
			expected: hexes.Hex_t{GridRow: 4, GridColumn: 19, Column: 30, Row: 1},
		},
		{
			name:     "Move SW from CS 3021",
			hex:      hexes.Hex_t{GridRow: 3, GridColumn: 19, Column: 30, Row: 21},
			dir:      direction.SouthWest,
			expected: hexes.Hex_t{GridRow: 4, GridColumn: 19, Column: 29, Row: 1},
		},
		{
			name:     "Move NW from CS 3021",
			hex:      hexes.Hex_t{GridRow: 3, GridColumn: 19, Column: 30, Row: 21},
			dir:      direction.NorthWest,
			expected: hexes.Hex_t{GridRow: 3, GridColumn: 19, Column: 29, Row: 21},
		},
		// Test moving from bottom left corner of a grid
		{
			name:     "Move N from CT 0121",
			hex:      hexes.Hex_t{GridRow: 3, GridColumn: 20, Column: 1, Row: 21},
			dir:      direction.North,
			expected: hexes.Hex_t{GridRow: 3, GridColumn: 20, Column: 1, Row: 20},
		},
		{
			name:     "Move NE from CT 0121",
			hex:      hexes.Hex_t{GridRow: 3, GridColumn: 20, Column: 1, Row: 21},
			dir:      direction.NorthEast,
			expected: hexes.Hex_t{GridRow: 3, GridColumn: 20, Column: 2, Row: 20},
		},
		{
			name:     "Move SE from CT 0121",
			hex:      hexes.Hex_t{GridRow: 3, GridColumn: 20, Column: 1, Row: 21},
			dir:      direction.SouthEast,
			expected: hexes.Hex_t{GridRow: 3, GridColumn: 20, Column: 2, Row: 21},
		},
		{
			name:     "Move S from CT 0121",
			hex:      hexes.Hex_t{GridRow: 3, GridColumn: 20, Column: 1, Row: 21},
			dir:      direction.South,
			expected: hexes.Hex_t{GridRow: 4, GridColumn: 20, Column: 1, Row: 1},
		},
		{
			name:     "Move SW from CT 0121",
			hex:      hexes.Hex_t{GridRow: 3, GridColumn: 20, Column: 1, Row: 21},
			dir:      direction.SouthWest,
			expected: hexes.Hex_t{GridRow: 3, GridColumn: 19, Column: 30, Row: 21},
		},
		{
			name:     "Move NW from CT 0121",
			hex:      hexes.Hex_t{GridRow: 3, GridColumn: 20, Column: 1, Row: 21},
			dir:      direction.NorthWest,
			expected: hexes.Hex_t{GridRow: 3, GridColumn: 19, Column: 30, Row: 20},
		},
		// Test moving from corners of the big map
		{
			name:     "Move NW from AA 0101",
			hex:      hexes.Hex_t{GridRow: 1, GridColumn: 26, Column: 1, Row: 1},
			dir:      direction.NorthEast,
			expected: hexes.Hex_t{GridRow: 26, GridColumn: 26, Column: 2, Row: 21},
		},
		{
			name:     "Move SE from ZZ 3021",
			hex:      hexes.Hex_t{GridRow: 26, GridColumn: 26, Column: 30, Row: 21},
			dir:      direction.SouthWest,
			expected: hexes.Hex_t{GridRow: 1, GridColumn: 26, Column: 29, Row: 1},
		},
		// Test moving from edges of a grid
		{
			name:     "Move N from AD 1701",
			hex:      hexes.Hex_t{GridRow: 1, GridColumn: 4, Column: 17, Row: 1},
			dir:      direction.North,
			expected: hexes.Hex_t{GridRow: 26, GridColumn: 4, Column: 17, Row: 21},
		},
		{
			name:     "Move S from ZD 1721",
			hex:      hexes.Hex_t{GridRow: 26, GridColumn: 4, Column: 17, Row: 21},
			dir:      direction.South,
			expected: hexes.Hex_t{GridRow: 1, GridColumn: 4, Column: 17, Row: 1},
		},
		{
			name:     "Move SE from PZ 3011",
			hex:      hexes.Hex_t{GridRow: 16, GridColumn: 26, Column: 30, Row: 11},
			dir:      direction.NorthEast,
			expected: hexes.Hex_t{GridRow: 16, GridColumn: 1, Column: 1, Row: 11},
		},
		{
			name:     "Move NW from PA 0110",
			hex:      hexes.Hex_t{GridRow: 16, GridColumn: 1, Column: 1, Row: 11},
			dir:      direction.NorthWest,
			expected: hexes.Hex_t{GridRow: 16, GridColumn: 26, Column: 30, Row: 10},
		},
	}

	for _, test := range tests {
		result := test.hex.Add(test.dir)
		if result != test.expected {
			t.Errorf("%s: want %q: got %q\n", test.name, test.expected, result)
		}
	}
}
