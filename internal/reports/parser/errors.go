// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package parser

// Error defines a constant error
type Error string

// Error implements the Errors interface
func (e Error) Error() string { return string(e) }

const (
	ErrNotASectionHeader = Error("not a section header")
)
