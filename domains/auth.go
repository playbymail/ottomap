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

// Claim_t is the type for a claim.
type Claim_t struct {
	UserId ID     // unique identifier for the user
	Clan   string // clan number, always 4 digits
	Roles  struct {
		IsActive        bool // true if the user is active
		IsAdministrator bool // true if the user is an administrator
		IsAuthenticated bool // true if the user is authenticated
		IsOperator      bool // true if the user is an operator
		IsUser          bool // true if the user is a user
	}
}

// Session_t is the type for a session.
type Session_t struct {
	Id        string    // unique identifier for the session
	CreatedAt time.Time // always UTC
	ExpiresAt time.Time // always UTC

	Claim_t // the claim for the session, embedded
}

// User_t is the type for a user.
type User_t struct {
	ID ID // unique identifier

	Email    string         // email address
	Timezone *time.Location // timezone, e.g. "America/New_York", should default to UTC

	HashedPassword string // bcrypt hashed password
	MagicLink      string // magic link, recommended to avoid using this

	Clan  string // clan number, always 4 digits
	Roles struct {
		IsActive        bool // true if the user is active
		IsAdministrator bool // true if the user is an administrator
		IsAuthenticated bool // true if the user is authenticated
		IsOperator      bool // true if the user is an operator
		IsUser          bool // true if the user is a user
	}

	Created   time.Time // always UTC
	Updated   time.Time // always UTC
	LastLogin time.Time // always UTC, time.Zero if never logged in
}

// authentication domain errors

var (
	ErrExpiredSession = errors.New("expired session")
	ErrInvalidClan    = errors.New("invalid clan")
	ErrInvalidEmail   = errors.New("invalid email")
	ErrInvalidHandle  = errors.New("invalid handle")
	ErrInvalidSession = errors.New("invalid session")
	ErrUnauthorized   = errors.New("unauthorized")
)
