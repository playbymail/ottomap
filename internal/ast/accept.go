// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package ast

import (
	"bytes"
	"unicode"
	"unicode/utf8"
)

// AcceptEOL returns the lexeme at the start of the input if it is an end of line.
// Returns nil, input if the input does not start with an end of line.
func AcceptEOL(input []byte) (lexeme, rest []byte) {
	offset := LenEOL(input)
	if offset == 0 {
		return nil, input
	}
	return []byte{'\n'}, input[offset:]
}

// AcceptInvalidRunes returns the run of invalid utf8 runes at the start of the input.
// Returns nil, input iIf the input does not start with an invalid utf8 rune.
func AcceptInvalidRunes(input []byte) (lexeme, rest []byte) {
	offset, buffer := 0, input
	for r, w := utf8.DecodeRune(buffer); len(buffer) != 0 && r == utf8.RuneError; r, w = utf8.DecodeRune(buffer) {
		offset, buffer = offset+w, buffer[w:]
	}
	if offset == 0 {
		return nil, input
	}
	return input[:offset], input[offset:]
}

// AcceptText returns the lexeme at the start of the input if it is text.
// Text is a run of input that is not EOL, invalid utf8, or whitespace.
// Returns nil, input if the input does not start with text.
func AcceptText(input []byte) (lexeme, rest []byte) {
	if len(input) == 0 || IsEOL(input) || IsWhitespace(input) {
		return nil, nil
	}

	// there are some symbols that are single characters. we will check for those first.
	ch := input[0]
	punctuation := []byte{'(', ')', '{', '}', '[', ']', ',', ';', ':', '=', '+', '-', '*', '/', '\\', '%', '<', '>', '!', '?', '&', '|', '^', '~', '#'}
	if bytes.IndexByte(punctuation, ch) != -1 {
		return []byte{ch}, input[1:]
	}

	// if we get to here, then the symbol is the run of characters up to the first EOL, whitespace, or invalid utf8.
	offset, buffer := 0, input
	for len(buffer) != 0 && !(IsEOL(buffer) || IsWhitespace(buffer)) {
		r, w := utf8.DecodeRune(buffer)
		if r == utf8.RuneError {
			break
		}
		offset, buffer = offset+w, buffer[w:]
	}
	if offset == 0 {
		return nil, input
	}
	return input[:offset], input[offset:]
}

// AcceptWhitespace returns the run of whitespace at the start of the input.
// Returns nil, input if the input does not start with whitespace.
// Note that end of line is not considered whitespace.
func AcceptWhitespace(input []byte) (lexeme, rest []byte) {
	offset, buffer := 0, input
	for r, w := utf8.DecodeRune(buffer); len(buffer) != 0 && unicode.IsSpace(r) && !IsEOL(buffer); r, w = utf8.DecodeRune(buffer) {
		offset, buffer = offset+w, buffer[w:]
	}
	if offset == 0 {
		return nil, input
	}
	return input[:offset], input[offset:]
}

func IsEOL(input []byte) bool {
	return LenEOL(input) != 0
}

func IsWhitespace(input []byte) bool {
	r, _ := utf8.DecodeRune(input)
	return unicode.IsSpace(r) && !IsEOL(input)
}

// LenEOL returns the length of the end of line at the start of the input.
// Returns 0 if the input does not start with an end of line.
func LenEOL(input []byte) int {
	if bytes.HasPrefix(input, []byte{'\n'}) { // unix line ending
		return 1
	} else if bytes.HasPrefix(input, []byte{'\r', '\n'}) { // windows line ending
		return 2
	} else if bytes.HasPrefix(input, []byte{'\r'}) { // mac line ending
		return 1
	}
	return 0
}
