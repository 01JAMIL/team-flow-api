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
SELECT count(*) OVER () AS total_count,
       id,
       workspace_name,
       description,
       user_id,
       created_at,
       updated_at
FROM workspaces
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

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

-- name: DeleteMemberFromWorkspace :exec
DELETE
FROM workspace_members
WHERE user_id = $1
  AND workspace_id = $2;

-- name: CreateProject :one
INSERT INTO projects (id, name, description, workspace_id)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetWorkspaceByID :one
SELECT *
FROM workspaces
WHERE id = $1::varchar;

-- name: GetWorkspaceProjects :many
SELECT count(*) OVER () AS total_count,
       id,
       name,
       description,
       workspace_id,
       created_at,
       updated_at
FROM projects
WHERE workspace_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetWorkspaceMembers :many
SELECT count(*) OVER () AS total_count,
       wm.id            AS member_id,
       wm.workspace_id,
       wm.user_role,
       wm.created_at    AS member_created_at,
       u.id::uuid       AS user_id,
       u.first_name,
       u.last_name,
       u.email,
       u.created_at,
       u.updated_at
FROM workspace_members wm
         JOIN users u ON u.id::uuid = wm.user_id::uuid
WHERE wm.workspace_id = $1
ORDER BY wm.created_at DESC
LIMIT $2 OFFSET $3;