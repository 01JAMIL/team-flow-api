-- +goose Up
ALTER TABLE integration_tasks
ADD COLUMN project_id UUID NOT NULL
REFERENCES projects (id)
ON DELETE CASCADE;

-- +goose Down
ALTER TABLE integration_tasks
DROP COLUMN IF EXISTS project_id;
