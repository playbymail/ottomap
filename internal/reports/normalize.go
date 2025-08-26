// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package reports

import "bytes"

var (
	// pre-computed lookup table for valid input characters
	isValid [256]bool

	// smart quotes and dashes replacements
	smartQuotesAndDashes = map[rune]byte{
		'\u2018': '\'', // left single quote
		'\u2019': '\'', // right single quote
		'\u201C': '"',  // left double quote
		'\u201D': '"',  // right double quote
		'\u2013': '-',  // en dash
		'\u2014': '-',  // em dash
		'\u2015': '-',  // horizontal bar
		'\u2212': '-',  // minus sign
	}
)

func init() {
	// initialize the lookup table for valid input characters
	for _, ch := range []byte(`abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ 0123456789 -,.'/\()`) {
		isValid[ch] = true
	}
}

// NormalizeReport does things and then returns the report as a slice of lines.
func NormalizeReport(input []byte) ([][]byte, error) {
	lines := bytes.Split(input, []byte{'\n'})

	for n, line := range lines {
		// replace smart quotes and dashes with ASCII equivalents
		runes := []rune(string(line))
		for i, r := range runes {
			if replacement, ok := smartQuotesAndDashes[r]; ok {
				runes[i] = rune(replacement)
			}
		}
		line = []byte(string(runes))

		// replace invalid characters with spaces
		for i, ch := range line {
			if !isValid[ch] {
				line[i] = ' '
			}
		}

		// trim trailing spaces
		line = bytes.TrimRight(line, " ")

		lines[n] = line
	}

	return lines, nil
}
