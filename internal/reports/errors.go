// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package reports

// Error defines a constant error
type Error string

// Error implements the Errors interface
func (e Error) Error() string { return string(e) }

const (
	ErrNoSections       = Error("no sections")
	ErrNotImplemented   = Error("not implemented")
	ErrTurnMissing      = Error("turn missing")
	ErrTurnDoesNotMatch = Error("turn does not match")
)
