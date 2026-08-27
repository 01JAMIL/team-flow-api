-- +goose Up
CREATE TABLE IF NOT EXISTS integration_tasks (
    id UUID PRIMARY KEY,
    provider TEXT NOT NULL CHECK (
        provider IN ('github', 'gitlab', 'jira')
    ),
    resource_type TEXT NOT NULL CHECK (
        resource_type IN ('issue','pull_request','ticket','merge_request')
    ),
    external_id TEXT NOT NULL UNIQUE,
    repository_name TEXT NOT NULL,
    issue_number INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL CHECK (
        status IN ('open', 'closed')
    ),
    assignee_id UUID REFERENCES users(id) ON DELETE SET NULL,
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS integration_tasks;