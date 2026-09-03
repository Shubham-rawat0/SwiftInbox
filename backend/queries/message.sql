-- name: GetMessages :many

SELECT
    id,
    sender,
    subject,
    created_at
FROM messages
WHERE mailbox_id = $1
ORDER BY created_at DESC
LIMIT 50;

-- name: GetRawMessage :one

SELECT
    id,
    sender,
    subject,
    created_at,
    raw
FROM messages
WHERE id = $1
  AND mailbox_id = $2;