// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package domains

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrCreateSchema        = Error("create schema")
	ErrCreateMeta          = Error("create metadata")
	ErrDatabaseExists      = Error("database exists")
	ErrForeignKeysDisabled = Error("foreign keys disabled")
	ErrInvalidPath         = Error("invalid path")
	ErrPragmaReturnedNil   = Error("pragma returned nil")
)
