--  Copyright (c) 2024 Michael D Henderson. All rights reserved.

-- GetSession returns the session with the given id.
-- Fails if the session does not exist or user is not active.
--
-- name: GetSession :one
SELECT users.user_id,
       users.email,
       users.timezone,
       users.is_active,
       users.is_administrator,
       users.is_operator,
       users.is_user,
       users.clan,
       users.created_at,
       users.updated_at,
       users.last_login,
       sessions.expires_at
FROM sessions
         INNER JOIN users ON sessions.user_id = users.user_id
WHERE sess_id = :session_id
  AND users.is_active = 1;

-- CreateUserSession creates a new session for the given user id.
--
-- name: CreateUserSession :exec
INSERT INTO sessions (sess_id, user_id, expires_at)
VALUES (:sess_id, :user_id, :expires_at);

-- DeleteUserSessions deletes all sessions for the given user id.
--
-- name: DeleteUserSessions :exec
DELETE
FROM sessions
WHERE user_id = :user_id;