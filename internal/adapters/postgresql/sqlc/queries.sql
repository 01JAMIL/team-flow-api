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
SELECT count(*) OVER () AS total_count, id,
       workspace_name,
       description,
       user_id,
       created_at,
       updated_at
FROM workspaces
WHERE user_id = $1
ORDER BY created_at DESC LIMIT $2
OFFSET $3;

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

-- name: GetProjectById :one
SELECT *
FROM projects
WHERE id = $1;

-- name: UpdateProject :one
UPDATE projects
SET name        = $2,
    description = $3,
    updated_at  = now()
WHERE id = $1 RETURNING *;

-- name: DeleteProject :exec
DELETE
FROM projects
WHERE id = $1;

-- name: GetWorkspaceByID :one
SELECT *
FROM workspaces
WHERE id = $1::varchar;

-- name: GetWorkspaceProjects :many
SELECT count(*) OVER () AS total_count, id,
       name,
       description,
       workspace_id,
       created_at,
       updated_at
FROM projects
WHERE workspace_id = $1
ORDER BY created_at DESC LIMIT $2
OFFSET $3;

-- name: GetWorkspaceMembers :many
SELECT count(*)         OVER () AS total_count, wm.id AS member_id,
       wm.workspace_id,
       wm.user_role,
       wm.created_at AS member_created_at,
       u.id::uuid       AS user_id, u.first_name,
       u.last_name,
       u.email,
       u.created_at,
       u.updated_at
FROM workspace_members wm
         JOIN users u ON u.id::uuid = wm.user_id::uuid
WHERE wm.workspace_id = $1
ORDER BY wm.created_at DESC
    LIMIT $2
OFFSET $3;

-- name: CreateTask :one
INSERT INTO tasks (id, name, description, start_date, end_date, status, priority, project_id, assignee_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;
-- name: GetProjectTasks :many
SELECT count(*) OVER () AS total_count, id,
       name,
       description,
       start_date,
       end_date,
       status,
       priority,
       project_id,
       assignee_id,
       created_at,
       updated_at
FROM tasks
WHERE project_id = $1
ORDER BY created_at DESC LIMIT $2
OFFSET $3;

-- name: GetTaskById :one
SELECT *
FROM tasks
WHERE id = $1;

-- name: DeleteTask :exec
DELETE
FROM tasks
WHERE id = $1;

-- name: UpdateTask :one
UPDATE tasks
SET name        = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    start_date  = COALESCE(sqlc.narg('start_date'), start_date),
    end_date    = COALESCE(sqlc.narg('end_date'), end_date),
    status      = COALESCE(sqlc.narg('status'), status),
    priority    = COALESCE(sqlc.narg('priority'), priority),
    assignee_id = COALESCE(sqlc.narg('assignee_id'), assignee_id),
    updated_at  = now()
WHERE id = sqlc.arg('id') RETURNING *;

-- name: CreateMessage :one
INSERT INTO messages (id, sender_id, receiver_id, content)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetMessagesBetweenUsers :many
SELECT count(*) OVER () AS total_count, id,
       sender_id,
       receiver_id,
       content,
       created_at
FROM messages
WHERE (sender_id = $1 AND receiver_id = $2)
   OR (sender_id = $2 AND receiver_id = $1)
ORDER BY created_at DESC LIMIT $3
OFFSET $4;

-- name: UpdateWorkspaceStripeCustomer :one
UPDATE workspaces
SET stripe_customer_id = $2
WHERE id = $1 RETURNING *;

-- name: CreateSubscription :one
INSERT INTO subscriptions (id,
                           workspace_id,
                           stripe_subscription_id,
                           stripe_price_id,
                           status,
                           plan,
                           current_period_start,
                           current_period_end)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;