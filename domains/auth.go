// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package domains

import (
	"errors"
	"time"
)

// ID is the type for identity. It is unique and immutable.
//
// It is used to identify a user, organization, or other entity.
//
// We assume that the ID is never deleted or reused.
type ID int64

// User is the type for a user.
type User struct {
	ID ID // unique identifier

	Handle   string         // username, not the user's full name
	Email    string         // email address
	Timezone *time.Location // timezone, e.g. "America/New_York", should default to UTC

	IsActive       bool   // true if the user is active
	HashedPassword string // hashed password, recommended to use bcrypt

	Clan  string
	Roles map[string]bool

	Created   time.Time // always UTC
	Updated   time.Time // always UTC
	LastLogin time.Time // always UTC, time.Zero if never logged in
}

// authentication domain errors

var (
	ErrInvalidEmail  = errors.New("invalid email")
	ErrInvalidHandle = errors.New("invalid handle")
	ErrInvalidClan   = errors.New("invalid clan")
	ErrUnauthorized  = errors.New("unauthorized")
)
