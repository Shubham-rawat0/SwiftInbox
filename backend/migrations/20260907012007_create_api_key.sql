-- +goose Up
CREATE TABLE IF NOT EXISTS apikeys (
    id  UUID PRIMARY KEY,
    developer_id  UUID NOT NULL REFERENCES developer(id) ON DELETE CASCADE,
    key_hash  TEXT  NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ
);

-- +goose Down
DROP TABLE apikeys;