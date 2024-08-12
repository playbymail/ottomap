// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package cerrs implements constant errors.
package cerrs

// Error defines a constant error
type Error string

// Error implements the Errors interface
func (e Error) Error() string { return string(e) }

const (
	ErrCreateMeta                 = Error("create metadata")
	ErrCreateSchema               = Error("create schema")
	ErrDatabaseExists             = Error("database exists")
	ErrDuplicateChecksum          = Error("duplicate checksum")
	ErrEmptyReport                = Error("empty report")
	ErrForeignKeysDisabled        = Error("foreign keys disabled")
	ErrInvalidGridCoordinates     = Error("invalid grid coordinates")
	ErrInvalidIndexFile           = Error("invalid index file")
	ErrInvalidInputPath           = Error("invalid input path")
	ErrInvalidOutputPath          = Error("invalid output path")
	ErrInvalidPath                = Error("invalid path")
	ErrInvalidReportFile          = Error("invalid report file")
	ErrMissingFollowsUnit         = Error("missing follows unit")
	ErrMissingIndexFile           = Error("missing index file")
	ErrMissingMovementResults     = Error("missing movement results")
	ErrMissingReportFile          = Error("missing report file")
	ErrMissingStatusLine          = Error("missing status line")
	ErrMultipleFleetMovementLines = Error("multiple fleet movement lines")
	ErrMultipleFollowsLines       = Error("multiple follows lines")
	ErrMultipleMovementLines      = Error("multiple movement lines")
	ErrMultipleStatusLines        = Error("multiple status lines")
	ErrNoSeparator                = Error("no separator")
	ErrNotAFile                   = Error("not a file")
	ErrNotAFleetMovementLine      = Error("not a fleet movement line")
	ErrNotATurnReport             = Error("not a turn report")
	ErrNotDirectory               = Error("not a directory")
	ErrNotImplemented             = Error("not implemented")
	ErrNotMovementResults         = Error("not movement results")
	ErrParseFailed                = Error("parse failed")
	ErrPragmaReturnedNil          = Error("pragma returned nil")
	ErrSetupExists                = Error("setup.json exists")
	ErrTooManyScoutLines          = Error("too many scout lines")
	ErrTrackingGarrison           = Error("tracking garrison")
	ErrUnableToFindStartingHex    = Error("unable to find starting hex")
	ErrUnexpectedNumberOfMoves    = Error("unexpected number of moves")
	ErrUnitMovesAndFollows        = Error("unit moves and follows")
)
