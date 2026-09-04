-- name: GetMessages :many

SELECT
    id,
    sender,
    subject,
    raw,
    created_at
FROM messages
WHERE address = $1
ORDER BY created_at DESC
LIMIT 50;

-- name: GetMessagesByMailboxID :many

SELECT
    id,
    sender,
    subject,
    raw,
    created_at
FROM messages
WHERE mailbox_id = $1
ORDER BY created_at DESC
LIMIT 50;

-- name: CreateMessage :one

INSERT INTO messages (
    id,
    mailbox_id,
    address,
    sender,
    subject,
    raw,
    expires_at
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING *;