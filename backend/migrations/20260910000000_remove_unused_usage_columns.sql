-- +goose Up
ALTER TABLE developer_usage
    DROP COLUMN IF EXISTS mailbox_requests,
    DROP COLUMN IF EXISTS message_requests;

-- +goose Down
ALTER TABLE developer_usage
    ADD COLUMN IF NOT EXISTS mailbox_requests INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS message_requests INT NOT NULL DEFAULT 0;