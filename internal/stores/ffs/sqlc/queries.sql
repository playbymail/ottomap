--  Copyright (c) 2024 Michael D Henderson. All rights reserved.


-- name: CreateUser :one
INSERT INTO users (handle, hashed_password, clan, magic_key, path)
VALUES (:handle, :hashed_password, :clan, :magic_key, :path)
RETURNING id;

-- name: CreateTurnReport :one
INSERT INTO reports (uid, turn, clan, path)
VALUES (:uid, :turn, :clan, :path)
RETURNING id;

-- name: CreateTurnMap :one
INSERT INTO maps (uid, turn, clan, path)
VALUES (:uid, :turn, :clan, :path)
RETURNING id;

-- name: CreateUnit :exec
INSERT INTO units (rid, turn, name, starting_hex, ending_hex)
VALUES (:rid, :turn, :name, :starting_hex, :ending_hex);

-- name: AuthenticateUser :one
SELECT id, clan
FROM users
WHERE clan = :clan
  AND hashed_password = :hashed_password;

-- name: GetUser :one
SELECT id, clan
FROM users
WHERE id = :id;

-- name: GetClan :one
SELECT clan
FROM users
WHERE id = :id;

-- name: GetUserReports :many
SELECT id, turn, clan, path
FROM reports
WHERE uid = :uid
ORDER BY turn, clan;

-- name: CreateSession :exec
INSERT INTO sessions (id, uid, expires_dttm)
VALUES (:id, :uid, :expires_dttm);

-- name: DeleteSession :exec
DELETE
FROM sessions
WHERE id = :id
   OR CURRENT_TIMESTAMP >= expires_dttm;

-- name: GetSession :one
SELECT id, uid, expires_dttm
FROM sessions
WHERE id = :id
  AND CURRENT_TIMESTAMP < expires_dttm;

