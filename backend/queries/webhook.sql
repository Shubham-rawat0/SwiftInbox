-- name: CreateWebhook :one
INSERT INTO webhooks(id,developer_id,url,secret_encrypted,events) values ($1,$2,$3,$4,$5) RETURNING id,developer_id,url,is_active,events;

-- name: GetWebhook :one
SELECT id, developer_id, url, secret_encrypted, is_active, events
FROM webhooks
WHERE id = $1 AND developer_id = $2;

-- name: ListWebhooks :many
SELECT id, developer_id, url, is_active, events
FROM webhooks
WHERE developer_id = $1
ORDER BY created_at DESC;

-- name: DeleteWebhook :one
DELETE FROM webhooks
WHERE id = $1 AND developer_id = $2
RETURNING id;