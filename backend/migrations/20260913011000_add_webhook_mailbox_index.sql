-- +goose Up
CREATE INDEX IF NOT EXISTS idx_webhook_mailboxes_mailbox_id
ON webhook_mailboxes (mailbox_id);

-- +goose Down
DROP INDEX IF EXISTS idx_webhook_mailboxes_mailbox_id;
