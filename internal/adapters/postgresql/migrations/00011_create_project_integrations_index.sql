-- +goose Up
CREATE INDEX idx_project_integrations_lookup
ON project_integrations (
    provider,
    repository_owner,
    repository_name
);

-- +goose Down
DROP INDEX IF EXISTS idx_project_integrations_lookup;
