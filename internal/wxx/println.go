// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

import "fmt"

// Printf is a convenience function to write a formatted text to the buffer.
func (w *WXX) Printf(format string, args ...any) {
	w.buffer.WriteString(fmt.Sprintf(format, args...))
}

// Println is a convenience function to write a formatted line to the buffer.
// it appends a newline to the buffer.
func (w *WXX) Println(format string, args ...any) {
	w.Printf(format, args...)
	w.buffer.WriteByte('\n')
}
