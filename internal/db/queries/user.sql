-- name: GetUser :one
SELECT id, username, created_at
FROM users
WHERE username = ?;

-- name: CreateUser :one
INSERT INTO users (
    id, username, created_at
) VALUES (
    ?, ?, ?
) RETURNING *;

-- name: GetUsers :many
SELECT id, username, created_at
FROM users;

-- name: DeleteUsersData :exec
DELETE FROM users;

