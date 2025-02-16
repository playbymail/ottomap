// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package norm

import "bytes"

var (
	// pre-computed lookup table for delimiters
	isSpaceDelimiter [256]bool
)

func init() {
	// initialize the lookup table for delimiters
	for _, ch := range []byte{'\n', ',', '(', ')', '\\', ':'} {
		isSpaceDelimiter[ch] = true
	}
}

// NormalizeSpaces returns a copy of the input with all runs of spaces and tabs replaced by a single space.
// Insignificant spaces (e.g., before and after delimiters) are removed.
// Example: "Tribe   0987 ,  ( status ) " -> "Tribe 0987,(status)".
func NormalizeSpaces(input []byte) []byte {
	if len(input) == 0 {
		return []byte{}
	}

	output := bytes.NewBuffer(make([]byte, 0, len(input)))
	prevCharWasDelimiter := false

	for i := 0; i < len(input); i++ {
		// skip runs of spaces and tabs
		if isSpaceOrTab(input[i]) {
			for ; i < len(input) && isSpaceOrTab(input[i]); i++ {
				// skip all spaces/tabs
			}
			// check if the space is significant or not
			nextCharIsDelimiter := i >= len(input) || isSpaceDelimiter[input[i]]
			isSignificantSpace := !(prevCharWasDelimiter || nextCharIsDelimiter)
			if isSignificantSpace {
				// space is significant, so keep it
				output.WriteByte(' ')
			}
			i-- // adjust for the outer loop increment
			continue
		}

		// write the current character and update the delimiter state
		output.WriteByte(input[i])
		prevCharWasDelimiter = isSpaceDelimiter[input[i]]
	}

	return output.Bytes()
}

// helper function to identify spaces and tabs
func isSpaceOrTab(b byte) bool {
	return b == ' ' || b == '\t'
}
