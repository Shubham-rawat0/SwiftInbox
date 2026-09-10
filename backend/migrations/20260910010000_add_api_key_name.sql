-- +goose Up
ALTER TABLE apikeys
    ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT 'default';

-- +goose Down
ALTER TABLE apikeys
    DROP COLUMN IF EXISTS name;