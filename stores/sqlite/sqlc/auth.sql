-- CreateUser creates a new user and returns its id.
-- The email must be lowercase and unique.
-- Timezone is the user's timezone. Use UTC if unknown.
-- The password is stored as a bcrypt hash.
-- Role should be one of "user", "admin", or "operator". Other values are accepted but ignored.
--
-- name: CreateUser :one
INSERT INTO users (email, timezone, is_active, hashed_password, clan, role, last_login)
VALUES (:email, :timezone, :is_active, :hashed_password, :clan, :role, :last_login)
RETURNING user_id;

-- AuthenticateUser authenticates a user with the given email address and password.
-- Returns the user's id if the authentication is successful.
-- Authentication will fail if the user is not active.
--
-- name: AuthenticateUser :one
SELECT user_id
FROM users
WHERE is_active = 1
  AND email = :email
  AND hashed_password = :hashed_password;


-- DeleteUser updates the user with the given id to inactive.
-- We do not delete the user because we want to keep the history of the user.
-- We also update the user's password and role to "deleted" to prevent them from logging in.
--
-- name: DeleteUser :exec
UPDATE users
SET is_active       = 0,
    hashed_password = 'deleted',
    role            = 'deleted',
    updated_at      = CURRENT_TIMESTAMP
WHERE user_id = :user_id;

-- DeleteUserByEmail updates the user with the given email address to inactive.
-- We do not delete the user because we want to keep the history of the user.
-- We also update the user's password and role to "deleted" to prevent them from logging in.
--
-- name: DeleteUserByEmail :exec
UPDATE users
SET is_active       = 0,
    hashed_password = 'deleted',
    role            = 'deleted',
    updated_at      = CURRENT_TIMESTAMP
WHERE email = :email;

-- GetUser returns the user with the given id.
-- Fails if the user is not active.
--
-- name: GetUser :one
SELECT email, timezone, is_active, clan, role
FROM users
WHERE is_active = 1
  AND user_id = :user_id;

-- GetUserByEmail returns the user id and hashed password for the given email address.
-- Fails if the user is not active.
--
-- name: GetUserByEmail :one
SELECT user_id, hashed_password
FROM users
WHERE is_active = 1
  AND email = :email;

-- UpdateUserLastLogin updates the last login time for the given user.
--
-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login = CURRENT_TIMESTAMP
WHERE user_id = :user_id;

-- UpdateUserPassword updates the password for the given user.
-- Fails if the user is not active.
--
-- name: UpdateUserPassword :exec
UPDATE users
SET hashed_password = :hashed_password,
    updated_at      = CURRENT_TIMESTAMP
WHERE is_active = 1
  AND user_id = :user_id;

-- UpdateUserTimezone updates the timezone for the given user.
--
-- name: UpdateUserTimezone :exec
UPDATE users
SET timezone   = :timezone,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = :user_id;
