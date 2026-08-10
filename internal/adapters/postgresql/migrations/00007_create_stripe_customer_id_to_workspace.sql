-- +goose Up
ALTER TABLE workspaces
ADD COLUMN stripe_customer_id VARCHAR(255);

-- +goose Down
ALTER TABLE workspaces
DROP COLUMN stripe_customer_id;
