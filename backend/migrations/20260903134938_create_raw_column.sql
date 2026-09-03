-- +goose Up
ALTER TABLE messages 
    DROP COLUMN message,
    ADD COLUMN raw BYTEA;
-- +goose Down
ALTER TABLE messages 
    DROP COLUMN raw;
