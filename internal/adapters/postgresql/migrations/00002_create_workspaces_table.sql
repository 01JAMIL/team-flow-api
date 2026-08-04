-- +goose Up
CREATE TABLE IF NOT EXISTS workspaces (
    id UUID PRIMARY KEY,
    workspace_name VARCHAR(255) NOT NULL,
    description VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT NULL,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)
);

-- +goose Down
DROP TABLE IF EXISTS workspaces;
