-- +goose Up
CREATE TABLE webhooks (
    id UUID PRIMARY KEY,
    developer_id UUID NOT NULL REFERENCES developer(id),
    url TEXT NOT NULL,
    secret_encrypted TEXT NOT NULL,
    events TEXT[] NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (developer_id, url)
);
-- +goose Down
DROP TABLE webhooks;
