-- +goose Up
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description VARCHAR(255) NOT NULL,
    workspace_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT NULL,
    CONSTRAINT fk_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id)
);

-- +goose Down
DROP TABLE IF EXISTS projects;
