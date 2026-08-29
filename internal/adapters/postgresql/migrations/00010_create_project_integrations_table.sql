-- +goose Up
CREATE TABLE IF NOT EXISTS project_integrations (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    provider TEXT NOT NULL CHECK (
        provider IN ('github', 'gitlab', 'jira')
    ),
    repository_owner VARCHAR(255) NOT NULL,
    repository_name VARCHAR(255) NOT NULL,
    webhook_secret VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT project_integrations_provider_repo_unique
    UNIQUE (
        provider,
        repository_owner,
        repository_name
    )
);

-- +goose Down
DROP TABLE IF EXISTS project_integrations;