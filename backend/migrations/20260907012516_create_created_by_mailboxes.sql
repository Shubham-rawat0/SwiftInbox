-- +goose Up
ALTER TABLE mailboxes ADD COLUMN created_by UUID;

-- +goose Down
ALTER TABLE mailboxes DROP COLUMN created_by;
