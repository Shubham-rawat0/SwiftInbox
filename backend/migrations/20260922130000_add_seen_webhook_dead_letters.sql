-- +goose Up
ALTER TABLE webhook_dead_letters ADD COLUMN seen BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE webhook_dead_letters DROP COLUMN seen;