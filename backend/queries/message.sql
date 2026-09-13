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

-- name: GetMessage :one

SELECT
    id,
    sender,
    subject,
    raw,
    address,
    created_at
FROM messages
WHERE id = $1;

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

-- name: DeleteExpiredMessages :one
WITH deleted AS (
    DELETE FROM messages
    WHERE expires_at < $1
    RETURNING id
)
SELECT COUNT(*)::INT AS deleted_count
FROM deleted;