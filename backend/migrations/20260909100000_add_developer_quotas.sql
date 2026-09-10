-- +goose Up
ALTER TABLE developer
    ADD COLUMN IF NOT EXISTS api_quota INT NOT NULL DEFAULT 1000,
    ADD COLUMN IF NOT EXISTS mailbox_quota INT NOT NULL DEFAULT 100,
    ADD COLUMN IF NOT EXISTS message_quota INT NOT NULL DEFAULT 1000;

-- +goose Down
ALTER TABLE developer
    DROP COLUMN IF EXISTS api_quota,
    DROP COLUMN IF EXISTS mailbox_quota,
    DROP COLUMN IF EXISTS message_quota;