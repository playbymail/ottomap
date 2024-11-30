// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package sqlite

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrCreateSchema        = Error("create schema")
	ErrDatabaseExists      = Error("database exists")
	ErrDuplicateHash       = Error("duplicate hash")
	ErrDuplicateReportName = Error("duplicate report name")
	ErrForeignKeysDisabled = Error("foreign keys disabled")
	ErrInvalidClanId       = Error("invalid clan id")
	ErrInvalidHash         = Error("invalid hash")
	ErrInvalidPath         = Error("invalid path")
	ErrInvalidMonth        = Error("invalid month")
	ErrInvalidUnit         = Error("invalid unit")
	ErrInvalidYear         = Error("invalid year")
	ErrNotDirectory        = Error("not a directory")
	ErrNotFound            = Error("not found")
	ErrPragmaReturnedNil   = Error("pragma returned nil")
)
