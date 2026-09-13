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

-- name: GetWebhooksByMailboxID :many
SELECT w.id, w.developer_id, w.url, w.secret_encrypted, w.is_active, w.events
FROM webhooks w
JOIN webhook_mailboxes wm ON wm.webhook_id = w.id
WHERE wm.mailbox_id = $1
ORDER BY w.created_at DESC;

-- name: DeleteWebhook :one
DELETE FROM webhooks
WHERE id = $1 AND developer_id = $2
RETURNING id;

-- name: LinkWebhookMailbox :one
INSERT INTO webhook_mailboxes(webhook_id, mailbox_id)
SELECT $1, m.id
FROM mailboxes m
JOIN webhooks w ON w.id = $1 AND w.developer_id = $3
WHERE m.id = $2 AND m.created_by = $3
ON CONFLICT (webhook_id, mailbox_id) DO NOTHING
RETURNING mailbox_id;

-- name: RemoveWebhookMailbox :one
DELETE FROM webhook_mailboxes wm
USING webhooks w
WHERE wm.webhook_id = w.id
  AND w.id = $1
  AND w.developer_id = $3
  AND wm.mailbox_id = $2
RETURNING wm.mailbox_id;

-- name: InsertEvent :one
UPDATE webhooks
SET events = ARRAY(
    SELECT DISTINCT unnest(events || $3)
)
WHERE id = $1
  AND developer_id = $2
RETURNING id, developer_id, url, is_active, events;

-- name: DeleteEvents :one
UPDATE webhooks
SET events = ARRAY(
  SELECT unnest(events)
  EXCEPT
  SELECT unnest($3::text[])
)
WHERE id = $1
  AND developer_id = $2
RETURNING id, developer_id, url, is_active, events;

-- name: CreateWebhookDeadLetter :exec
INSERT INTO webhook_dead_letters (
    id, developer_id, webhook_id, mailbox_id, message_id,
    event, url, reason, attempts
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: ListWebhookDeadLetters :many
SELECT id, webhook_id, mailbox_id, message_id, event, url, reason, attempts, created_at
FROM webhook_dead_letters
WHERE developer_id = $1
ORDER BY created_at DESC
LIMIT 100;