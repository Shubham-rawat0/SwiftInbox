-- +goose Up
CREATE TABLE IF NOT EXISTS api_key_usage (
    api_key_id UUID NOT NULL REFERENCES apikeys(id) ON DELETE CASCADE,
    period DATE NOT NULL,
    api_requests INT NOT NULL DEFAULT 0,
    mailbox_requests INT NOT NULL DEFAULT 0,
    message_requests INT NOT NULL DEFAULT 0,

    PRIMARY KEY (api_key_id, period)
);

-- +goose Down
DROP TABLE IF EXISTS api_key_usage;