// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: auth.sql

package sqlc

import (
	"context"
	"time"
)

const authenticateUser = `-- name: AuthenticateUser :one
SELECT user_id
FROM users
WHERE is_active = 1
  AND email = ?1
  AND hashed_password = ?2
`

type AuthenticateUserParams struct {
	Email          string
	HashedPassword string
}

// AuthenticateUser authenticates a user with the given email address and password.
// Returns the user's id if the authentication is successful.
// Authentication will fail if the user is not active.
func (q *Queries) AuthenticateUser(ctx context.Context, arg AuthenticateUserParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, authenticateUser, arg.Email, arg.HashedPassword)
	var user_id int64
	err := row.Scan(&user_id)
	return user_id, err
}

const createOperator = `-- name: CreateOperator :exec
INSERT INTO users (email, timezone, is_active, hashed_password, clan, role, last_login)
VALUES ('operator', 'UTC', 1, ?1, '0000', 'operator', '')
ON CONFLICT (email) DO UPDATE SET is_active       = 1,
                                  hashed_password = ?1
`

// CreateOperator creates a new operator or updates an existing one.
func (q *Queries) CreateOperator(ctx context.Context, hashedPassword string) error {
	_, err := q.db.ExecContext(ctx, createOperator, hashedPassword)
	return err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (email, timezone, is_active, hashed_password, clan, role, last_login)
VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7)
RETURNING user_id
`

type CreateUserParams struct {
	Email          string
	Timezone       string
	IsActive       int64
	HashedPassword string
	Clan           string
	Role           string
	LastLogin      time.Time
}

// CreateUser creates a new user and returns its id.
// The email must be lowercase and unique.
// Timezone is the user's timezone. Use UTC if unknown.
// The password is stored as a bcrypt hash.
// Role should be one of "user", "admin", or "operator". Other values are accepted but ignored.
func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.Email,
		arg.Timezone,
		arg.IsActive,
		arg.HashedPassword,
		arg.Clan,
		arg.Role,
		arg.LastLogin,
	)
	var user_id int64
	err := row.Scan(&user_id)
	return user_id, err
}

const deleteUser = `-- name: DeleteUser :exec
UPDATE users
SET is_active       = 0,
    hashed_password = 'deleted',
    role            = 'deleted',
    updated_at      = CURRENT_TIMESTAMP
WHERE user_id = ?1
`

// DeleteUser updates the user with the given id to inactive.
// We do not delete the user because we want to keep the history of the user.
// We also update the user's password and role to "deleted" to prevent them from logging in.
func (q *Queries) DeleteUser(ctx context.Context, userID int64) error {
	_, err := q.db.ExecContext(ctx, deleteUser, userID)
	return err
}

const deleteUserByEmail = `-- name: DeleteUserByEmail :exec
UPDATE users
SET is_active       = 0,
    hashed_password = 'deleted',
    role            = 'deleted',
    updated_at      = CURRENT_TIMESTAMP
WHERE email = ?1
`

// DeleteUserByEmail updates the user with the given email address to inactive.
// We do not delete the user because we want to keep the history of the user.
// We also update the user's password and role to "deleted" to prevent them from logging in.
func (q *Queries) DeleteUserByEmail(ctx context.Context, email string) error {
	_, err := q.db.ExecContext(ctx, deleteUserByEmail, email)
	return err
}

const getUser = `-- name: GetUser :one
SELECT email, timezone, is_active, clan, role
FROM users
WHERE is_active = 1
  AND user_id = ?1
`

type GetUserRow struct {
	Email    string
	Timezone string
	IsActive int64
	Clan     string
	Role     string
}

// GetUser returns the user with the given id.
// Fails if the user is not active.
func (q *Queries) GetUser(ctx context.Context, userID int64) (GetUserRow, error) {
	row := q.db.QueryRowContext(ctx, getUser, userID)
	var i GetUserRow
	err := row.Scan(
		&i.Email,
		&i.Timezone,
		&i.IsActive,
		&i.Clan,
		&i.Role,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT user_id, hashed_password
FROM users
WHERE is_active = 1
  AND email = ?1
`

type GetUserByEmailRow struct {
	UserID         int64
	HashedPassword string
}

// GetUserByEmail returns the user id and hashed password for the given email address.
// Fails if the user is not active.
func (q *Queries) GetUserByEmail(ctx context.Context, email string) (GetUserByEmailRow, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var i GetUserByEmailRow
	err := row.Scan(&i.UserID, &i.HashedPassword)
	return i, err
}

const updateUserLastLogin = `-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login = CURRENT_TIMESTAMP
WHERE user_id = ?1
`

// UpdateUserLastLogin updates the last login time for the given user.
func (q *Queries) UpdateUserLastLogin(ctx context.Context, userID int64) error {
	_, err := q.db.ExecContext(ctx, updateUserLastLogin, userID)
	return err
}

const updateUserPassword = `-- name: UpdateUserPassword :exec
UPDATE users
SET hashed_password = ?1,
    updated_at      = CURRENT_TIMESTAMP
WHERE is_active = 1
  AND user_id = ?2
`

type UpdateUserPasswordParams struct {
	HashedPassword string
	UserID         int64
}

// UpdateUserPassword updates the password for the given user.
// Fails if the user is not active.
func (q *Queries) UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) error {
	_, err := q.db.ExecContext(ctx, updateUserPassword, arg.HashedPassword, arg.UserID)
	return err
}

const updateUserTimezone = `-- name: UpdateUserTimezone :exec
UPDATE users
SET timezone   = ?1,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = ?2
`

type UpdateUserTimezoneParams struct {
	Timezone string
	UserID   int64
}

// UpdateUserTimezone updates the timezone for the given user.
func (q *Queries) UpdateUserTimezone(ctx context.Context, arg UpdateUserTimezoneParams) error {
	_, err := q.db.ExecContext(ctx, updateUserTimezone, arg.Timezone, arg.UserID)
	return err
}