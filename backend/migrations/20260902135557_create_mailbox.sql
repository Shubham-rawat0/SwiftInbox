-- +goose Up
CREATE TABLE IF NOT EXISTS mailboxes (
    id UUID PRIMARY KEY,
    address TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_mailboxes_expires_at
ON mailboxes (expires_at);

-- +goose Down
DROP TABLE IF EXISTS mailboxes;
