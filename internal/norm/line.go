// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package norm

import "regexp"

var (
	// pre-computed lookup table for valid input characters
	isValid [256]bool

	// replacement expressions
	reBackslashDash  = regexp.MustCompile(`\\+-+ *`)
	reBackslashUnit  = regexp.MustCompile(`\\+ *(\d{4}(?:[cefg]\d)?)`)
	reDashSpacesUnit = regexp.MustCompile(`-( *\d{4}([cefg]\d)?)`)
	reDirectionUnit  = regexp.MustCompile(`\b(NE|SE|SW|NW|N|S) +(\d{4}(?:[cefg]\d)?)`)
)

func init() {
	// initialize the lookup table for valid input characters
	for _, ch := range []byte(`abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ 0123456789 -,.'\`) {
		isValid[ch] = true
	}
}

func NormalizeLine(input []byte) []byte {
	if len(input) == 0 {
		return []byte{}
	}

	// replace invalid characters with spaces
	line := make([]byte, 0, len(input))
	for _, ch := range input {
		if !isValid[ch] {
			ch = ' '
		}
		line = append(line, ch)
	}

	// replace backslash+dash with comma+space
	line = reBackslashDash.ReplaceAll(line, []byte{',', ' '})

	// replace dash + optionalSpaces + unit with comma+spaces+unit
	line = reDashSpacesUnit.ReplaceAll(line, []byte{',', '$', '1'})

	// fix issues with backslash followed by a unit ID
	line = reBackslashUnit.ReplaceAll(line, []byte{',', '$', '1'})

	// fix issues with direction followed by a unit ID
	line = reDirectionUnit.ReplaceAll(line, []byte{'$', '1', ',', '$', '2'})

	return line
}
