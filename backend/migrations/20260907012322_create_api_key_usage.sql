-- +goose Up
CREATE TABLE IF NOT EXISTS developer_usage (
    developer_id UUID NOT NULL REFERENCES developer(id) ON DELETE CASCADE,
    period DATE NOT NULL,
    api_requests INT NOT NULL DEFAULT 0,
    mailboxes_created INT NOT NULL DEFAULT 0,
    messages_received INT NOT NULL DEFAULT 0,

    PRIMARY KEY (developer_id, period)
);

-- +goose Down
DROP TABLE developer_usage;