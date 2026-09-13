-- +goose Up
CREATE TABLE webhook_dead_letters (
    id UUID PRIMARY KEY,
    developer_id UUID NOT NULL REFERENCES developer(id) ON DELETE CASCADE,
    webhook_id UUID NOT NULL,
    mailbox_id UUID,
    message_id UUID,
    event TEXT NOT NULL,
    url TEXT NOT NULL,
    reason TEXT NOT NULL,
    attempts INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_dead_letters_developer_created_at
ON webhook_dead_letters (developer_id, created_at DESC);

-- +goose Down
DROP TABLE webhook_dead_letters;