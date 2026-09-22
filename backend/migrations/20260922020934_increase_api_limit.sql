-- +goose Up

ALTER TABLE developer
    ALTER COLUMN api_quota SET DEFAULT 5000,
    ALTER COLUMN mailbox_quota SET DEFAULT 100,
    ALTER COLUMN message_quota SET DEFAULT 1000;

-- +goose Down

ALTER TABLE developer
    ALTER COLUMN api_quota SET DEFAULT 1000,
    ALTER COLUMN mailbox_quota SET DEFAULT 100,
    ALTER COLUMN message_quota SET DEFAULT 1000;