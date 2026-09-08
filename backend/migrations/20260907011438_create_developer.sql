-- +goose Up
CREATE TABLE IF NOT EXISTS developer (
    id  UUID PRIMARY KEY,
    name   TEXT  NOT NULL,
    email  TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- +goose Down
DROP TABLE developer;
