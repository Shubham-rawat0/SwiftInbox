-- name: CreateWebhook :one
INSERT INTO webhooks(id,developer_id,name,url,secret_encrypted,events) values ($1,$2,$3,$4,$5,$6) RETURNING id,developer_id,name,url,is_active,events;

-- name: GetWebhook :one
SELECT w.id, w.developer_id, w.name, w.url, w.secret_encrypted, w.is_active, w.events,
    COALESCE(
        ARRAY(SELECT wm.mailbox_id::uuid FROM webhook_mailboxes wm WHERE wm.webhook_id = w.id ORDER BY wm.mailbox_id::uuid),
        '{}'::uuid[]
    )::uuid[] AS mailbox_ids
FROM webhooks w
WHERE w.id = $1 AND w.developer_id = $2;

-- name: ListWebhooks :many
SELECT w.id, w.developer_id, w.name, w.url, w.is_active, w.events,
    COALESCE(
        ARRAY(SELECT wm.mailbox_id::uuid FROM webhook_mailboxes wm WHERE wm.webhook_id = w.id ORDER BY wm.mailbox_id::uuid),
        '{}'::uuid[]
    )::uuid[] AS mailbox_ids
FROM webhooks w
WHERE w.developer_id = $1
ORDER BY w.created_at DESC;

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

-- name: UpdateWebhook :one
UPDATE webhooks w
SET name = COALESCE(NULLIF($3::text, ''), name),
    is_active = $4::boolean,
    updated_at = CURRENT_TIMESTAMP
WHERE w.id = $1 AND w.developer_id = $2
RETURNING w.id, w.developer_id, w.name, w.url, w.is_active, w.events,
    COALESCE(
        ARRAY(SELECT wm.mailbox_id::uuid FROM webhook_mailboxes wm WHERE wm.webhook_id = w.id ORDER BY wm.mailbox_id::uuid),
        '{}'::uuid[]
    )::uuid[] AS mailbox_ids;

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

-- name: GetWebhookMailbox :one
SELECT wm.webhook_id, wm.mailbox_id
FROM webhook_mailboxes wm
JOIN webhooks w ON w.id = wm.webhook_id AND w.developer_id = $3
WHERE wm.webhook_id = $1 AND wm.mailbox_id = $2;

-- name: InsertEvent :one
UPDATE webhooks w
SET events = ARRAY(
    SELECT DISTINCT unnest(events || $3::text[])
)
WHERE w.id = $1
  AND w.developer_id = $2
RETURNING w.id, w.developer_id, w.name, w.url, w.is_active, w.events,
    COALESCE(
        ARRAY(SELECT wm.mailbox_id::uuid FROM webhook_mailboxes wm WHERE wm.webhook_id = w.id ORDER BY wm.mailbox_id::uuid),
        '{}'::uuid[]
    )::uuid[] AS mailbox_ids;

-- name: DeleteEvents :one
UPDATE webhooks w
SET events = ARRAY(
  SELECT unnest(events)
  EXCEPT
  SELECT unnest($3::text[])
)
WHERE w.id = $1
  AND w.developer_id = $2
RETURNING w.id, w.developer_id, w.name, w.url, w.is_active, w.events,
    COALESCE(
        ARRAY(SELECT wm.mailbox_id::uuid FROM webhook_mailboxes wm WHERE wm.webhook_id = w.id ORDER BY wm.mailbox_id::uuid),
        '{}'::uuid[]
    )::uuid[] AS mailbox_ids;

-- name: CreateWebhookDeadLetter :exec
INSERT INTO webhook_dead_letters (
    id, developer_id, webhook_id, mailbox_id, message_id,
    event, url, reason, attempts
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: ListWebhookDeadLetters :many
SELECT d.id, d.webhook_id, d.mailbox_id, d.message_id, d.event, d.url, d.reason, d.attempts, d.created_at, d.seen,
    m.address AS mailbox_address,
    msg.sender AS message_sender,
    msg.subject AS message_subject
FROM webhook_dead_letters d
LEFT JOIN mailboxes m ON m.id = d.mailbox_id
LEFT JOIN messages msg ON msg.id = d.message_id
WHERE d.developer_id = $1
ORDER BY d.created_at DESC
LIMIT 100;

-- name: MarkWebhookDeadLetterSeen :one
UPDATE webhook_dead_letters
SET seen = TRUE
WHERE id = $1 AND developer_id = $2
RETURNING id;
