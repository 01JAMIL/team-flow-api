-- +goose Up
CREATE TABLE IF NOT EXISTS workspace_members (
    id UUID PRIMARY KEY,
    user_id VARCHAR(255),
    workspace_id VARCHAR(255),
    user_role TEXT NOT NULL CHECK (user_role IN ('ADMIN', 'MEMBER')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id)
);

-- +goose Down
DROP TABLE IF EXISTS workspace_members;