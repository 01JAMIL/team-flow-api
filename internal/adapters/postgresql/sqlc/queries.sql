/* Authentication */

-- name: Register :one
INSERT INTO users (id, first_name, last_name, email, password)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetUserById :one
SELECT *
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;

/* Workspace */

-- name: GetUserWorkspaceByID :one
SELECT *
FROM workspaces
WHERE user_id = $1
  AND id = $2;

-- name: GetUserWorkspaces :many
SELECT *
FROM workspaces
WHERE user_id = $1;