-- +goose Up
ALTER TABLE developer_usage
RENAME COLUMN mailboxes_created TO mailbox_requests;
ALTER TABLE developer_usage
RENAME COLUMN messages_received TO messages_requests;


-- +goose Down
ALTER TABLE developer_usage
RENAME COLUMN mailbox_requests TO mailboxes_created;
ALTER TABLE developer_usage
RENAME COLUMN messages_requests TO messages_received;
