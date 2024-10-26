// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package ast_test

import (
	"bytes"
	"github.com/playbymail/ottomap/internal/ast"
	"testing"
)

func TestAcceptEOL(t *testing.T) {
	tests := []struct {
		id     int
		input  []byte
		lexeme []byte
		rest   []byte
	}{
		{1, []byte("\nHello"), []byte("\n"), []byte("Hello")},   // Unix EOL
		{2, []byte("\r\nHello"), []byte("\n"), []byte("Hello")}, // Windows EOL
		{3, []byte("\rHello"), []byte("\n"), []byte("Hello")},   // Older Mac EOL
		{4, []byte("No EOL"), nil, []byte("No EOL")},            // No EOL
		{5, []byte(""), nil, []byte("")},                        // Empty string input
		{6, []byte{}, nil, []byte{}},                            // Empty input
		{7, nil, nil, []byte("")},                               // nil input
	}

	for _, tt := range tests {
		lexeme, rest := ast.AcceptEOL(tt.input)
		if !bytes.Equal(lexeme, tt.lexeme) || !bytes.Equal(rest, tt.rest) {
			t.Errorf("acceptEOL(%d): lexeme: want %q: got %q", tt.id, lexeme, tt.lexeme)
		} else if !bytes.Equal(rest, tt.rest) {
			t.Errorf("acceptEOL(%d): rest: want %q: got %q", tt.id, rest, tt.rest)
		}
	}
}

func TestAcceptInvalidRunes(t *testing.T) {
	tests := []struct {
		id     int
		input  []byte
		lexeme []byte
		rest   []byte
	}{
		{1, []byte{0xFF, 0xFF, 0x61}, []byte{0xFF, 0xFF}, []byte{0x61}}, // Invalid UTF-8 followed by valid 'a'
		{2, []byte{0xC3, 0x28}, []byte{0xC3}, []byte{0x28}},             // Invalid UTF-8 followed by valid '('
		{3, []byte("valid"), nil, []byte("valid")},                      // Valid UTF-8 input
		{4, []byte(""), nil, []byte("")},                                // Empty string input
		{5, []byte{}, nil, []byte{}},                                    // Empty input
		{6, nil, nil, []byte{}},                                         // nil input
	}

	for _, tt := range tests {
		lexeme, rest := ast.AcceptInvalidRunes(tt.input)
		if !bytes.Equal(lexeme, tt.lexeme) {
			t.Errorf("acceptInvalidRunes(%d): lexeme: want %q: got %q", tt.id, tt.lexeme, lexeme)
		} else if !bytes.Equal(rest, tt.rest) {
			t.Errorf("acceptInvalidRunes(%d): rest: want %q: got %q", tt.id, tt.rest, rest)
		}
	}
}

func TestAcceptText(t *testing.T) {
	tests := []struct {
		id     int
		input  []byte
		lexeme []byte
		rest   []byte
	}{
		{1, []byte("Hello"), []byte("Hello"), []byte{}},
		{2, []byte("Hello, World!"), []byte("Hello,"), []byte(" World!")},
		{3, []byte("$123.45 !"), []byte("$123.45"), []byte(" !")},
		{4, []byte("Sam,Fine"), []byte("Sam,Fine"), []byte{}},
		{5, []byte(",Fine"), []byte(","), []byte("Fine")},
		{6, []byte("(Previous Hex = ## 1234)"), []byte("("), []byte("Previous Hex = ## 1234)")},
		{7, []byte{'1', '2', 0xC3, 0x28}, []byte{'1', '2'}, []byte{0xC3, 0x28}},
		{8, []byte{}, nil, []byte{}}, // Empty input
		{9, nil, nil, []byte{}},      // nil input
	}

	for _, tt := range tests {
		lexeme, rest := ast.AcceptText(tt.input)
		if !bytes.Equal(lexeme, tt.lexeme) {
			t.Errorf("acceptSymbol(%d): lexeme: want %q: got %q", tt.id, tt.lexeme, lexeme)
		} else if !bytes.Equal(rest, tt.rest) {
			t.Errorf("acceptSymbol(%d): rest: want %q: got %q", tt.id, tt.rest, rest)
		}
	}
}

func TestAcceptWhitespace(t *testing.T) {
	tests := []struct {
		id     int
		input  []byte
		lexeme []byte
		rest   []byte
	}{
		{1, []byte("   Hello"), []byte("   "), []byte("Hello")},   // Spaces
		{2, []byte("\t\tHello"), []byte("\t\t"), []byte("Hello")}, // Tabs
		{3, []byte("Hello"), nil, []byte("Hello")},                // No whitespace
		{4, []byte("\nLinux"), nil, []byte("\nLinux")},            // EOL (shouldn't be considered whitespace)
		{5, []byte("\r\nWindows"), nil, []byte("\r\nWindows")},    // EOL (shouldn't be considered whitespace)
		{6, []byte("\rMacOS"), nil, []byte("\rMacOS")},            // EOL (shouldn't be considered whitespace)
		{7, []byte(""), nil, []byte("")},                          // Empty string input
		{8, []byte{}, nil, []byte{}},                              // Empty input
		{9, nil, nil, []byte{}},                                   // nil input
	}

	for _, tt := range tests {
		lexeme, rest := ast.AcceptWhitespace(tt.input)
		if !bytes.Equal(lexeme, tt.lexeme) {
			t.Errorf("acceptWhitespace(%d): lexeme: want %q: got %q", tt.id, tt.lexeme, lexeme)
		} else if !bytes.Equal(rest, tt.rest) {
			t.Errorf("acceptWhitespace(%d): rest: want %q: got %q", tt.id, tt.rest, rest)
		}
	}
}

func TestLenEol(t *testing.T) {
	tests := []struct {
		id       int
		input    []byte
		expected int
	}{
		{1, []byte("\r\nHello"), 2}, // Windows EOL
		{2, []byte("\nHello"), 1},   // Unix EOL
		{3, []byte("\rHello"), 1},   // Older Mac EOL
		{4, []byte("No EOL"), 0},    // No EOL
		{5, []byte(""), 0},          // Empty string input
		{6, []byte{}, 0},            // Empty input
		{7, nil, 0},                 // nil input
	}

	for _, tt := range tests {
		result := ast.LenEOL(tt.input)
		if result != tt.expected {
			t.Errorf("lenEol(%d) want %d, got %d", tt.input, tt.expected, result)
		}
	}
}
