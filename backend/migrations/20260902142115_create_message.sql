-- +goose Up

CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY,
    sender TEXT NOT NULL,
    message TEXT,
    subject TEXT,
    mailbox_id UUID NOT NULL REFERENCES mailboxes(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_expires_at
ON messages (expires_at);

-- +goose Down

DROP TABLE IF EXISTS messages;

