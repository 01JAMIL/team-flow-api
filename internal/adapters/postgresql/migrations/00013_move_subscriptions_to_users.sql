-- +goose Up
-- Move subscription ownership from workspace to user.
-- The Stripe customer also moves from the workspace to the user.

-- 1. Add stripe_customer_id to users (mirrors what workspaces had).
ALTER TABLE users
ADD COLUMN stripe_customer_id VARCHAR(255);

-- 2. Add user_id to subscriptions (nullable so existing rows can be backfilled).
ALTER TABLE subscriptions
ADD COLUMN user_id UUID REFERENCES users (id);

-- 3. Backfill user_id from the owner of the subscription's workspace.
UPDATE subscriptions s
SET user_id = w.user_id
FROM workspaces w
WHERE s.workspace_id = w.id;

-- 4. user_id is now populated; make it mandatory and index it.
ALTER TABLE subscriptions
ALTER COLUMN user_id SET NOT NULL;

CREATE INDEX idx_subscriptions_user_id ON subscriptions (user_id);

-- 5. Drop the workspace ownership column (drops the FK to workspaces too).
ALTER TABLE subscriptions
DROP COLUMN workspace_id;

-- 6. Move the existing Stripe customer from the user's workspace(s) to the user.
UPDATE users u
SET stripe_customer_id = w.stripe_customer_id
FROM workspaces w
WHERE u.stripe_customer_id IS NULL
  AND w.user_id = u.id
  AND w.stripe_customer_id IS NOT NULL;

-- 7. Drop the now-unused Stripe customer column from workspaces.
ALTER TABLE workspaces
DROP COLUMN stripe_customer_id;

-- +goose Down
-- 1. Add workspace_id back (nullable so rows can be re-attached).
ALTER TABLE subscriptions
ADD COLUMN workspace_id UUID REFERENCES workspaces (id);

-- 2. Re-attach each subscription to one of the user's workspaces (best effort).
UPDATE subscriptions s
SET workspace_id = w.id
FROM workspaces w
WHERE s.workspace_id IS NULL
  AND w.user_id = s.user_id;

-- 3. Restore the NOT NULL constraint and drop the user ownership column.
ALTER TABLE subscriptions
ALTER COLUMN workspace_id SET NOT NULL;

ALTER TABLE subscriptions
DROP COLUMN user_id;

-- 4. Put the Stripe customer back on the workspaces.
ALTER TABLE workspaces
ADD COLUMN stripe_customer_id VARCHAR(255);

UPDATE workspaces w
SET stripe_customer_id = u.stripe_customer_id
FROM users u
WHERE u.id = w.user_id
  AND u.stripe_customer_id IS NOT NULL;

-- 5. Drop the Stripe customer column from users.
ALTER TABLE users
DROP COLUMN stripe_customer_id;