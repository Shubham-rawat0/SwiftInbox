-- +goose Up
CREATE TABLE webhook_mailboxes (
    webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    mailbox_id UUID NOT NULL REFERENCES mailboxes(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (webhook_id, mailbox_id)
);

CREATE INDEX idx_webhook_mailboxes_mailbox_id
ON webhook_mailboxes (mailbox_id);

-- +goose Down
DROP TABLE webhook_mailboxes;