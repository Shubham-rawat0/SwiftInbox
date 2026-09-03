-- +goose Up
ALTER TABLE messages ADD COLUMN address TEXT NOT NULL;

-- +goose Down
ALTER TABLE messages DROP COLUMN address;
