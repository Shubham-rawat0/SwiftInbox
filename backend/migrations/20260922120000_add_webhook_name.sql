-- +goose Up
ALTER TABLE webhooks ADD COLUMN name TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE webhooks DROP COLUMN name;