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

-- name: GetUserWorkspaceByID :one
SELECT *
FROM workspaces
WHERE user_id = $1
  AND id = $2;

-- name: GetUserWorkspaces :many
SELECT *
FROM workspaces
WHERE user_id = $1;

-- name: CreateWorkspace :one
INSERT INTO workspaces (id, workspace_name, description, user_id)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateWorkspace :one
UPDATE workspaces
SET workspace_name = $3,
    description    = $4,
    updated_at     = now()
WHERE id = $1
  AND user_id = $2 RETURNING *;

-- name: DeleteWorkspace :exec
DELETE
FROM workspaces
WHERE id = $1
  AND user_id = $2;

-- name: GetWorkspaceById :one
SELECT *
FROM workspaces
WHERE id = $1;

-- name: AddWorkspaceMember :one
INSERT INTO workspace_members (id, user_id, workspace_id, user_role)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetMemberFromWorkspace :one
SELECT *
FROM workspace_members
WHERE user_id = $1
  AND workspace_id = $2;