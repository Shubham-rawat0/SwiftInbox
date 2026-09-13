-- +goose Up
ALTER TABLE developer
    ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE developer
    DROP COLUMN IF EXISTS password_hash;
