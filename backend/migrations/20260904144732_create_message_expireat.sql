-- +goose Up

ALTER TABLE messages
ALTER COLUMN expires_at SET NOT NULL;

-- +goose Down

ALTER TABLE messages
ALTER COLUMN expires_at DROP NOT NULL;