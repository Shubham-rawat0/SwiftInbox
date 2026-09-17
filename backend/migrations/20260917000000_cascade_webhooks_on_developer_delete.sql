-- +goose Up
ALTER TABLE webhooks
    DROP CONSTRAINT webhooks_developer_id_fkey,
    ADD CONSTRAINT webhooks_developer_id_fkey
        FOREIGN KEY (developer_id)
        REFERENCES developer(id)
        ON DELETE CASCADE;

-- +goose Down
ALTER TABLE webhooks
    DROP CONSTRAINT webhooks_developer_id_fkey,
    ADD CONSTRAINT webhooks_developer_id_fkey
        FOREIGN KEY (developer_id)
        REFERENCES developer(id);
