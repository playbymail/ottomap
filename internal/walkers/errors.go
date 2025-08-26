// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package walkers

// Error defines a constant error
type Error string

// Error implements the Errors interface
func (e Error) Error() string { return string(e) }

const (
	ErrNotImplemented = Error("not implemented")
)
