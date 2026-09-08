-- name: CreateEmailAddress :one
INSERT INTO mailboxes(
    id , address , expires_at, created_by
) 
VALUES ($1,$2,$3,$4)
ON CONFLICT(address)
DO UPDATE SET
    expires_at = EXCLUDED.expires_at,
    created_by = COALESCE(EXCLUDED.created_by, mailboxes.created_by)
RETURNING address, created_at, expires_at, created_by;

-- name: CreateCustomEmailAddress :one
INSERT INTO mailboxes(
    id, address, expires_at, created_by
)
VALUES ($1, $2, $3, $4)
ON CONFLICT(address) DO NOTHING
RETURNING address, created_at, expires_at, created_by;

-- name: UpsertMailbox :one

INSERT INTO mailboxes (
    id,
    address,
    expires_at,
    created_by
)
VALUES (
    $1,
    $2,
    $3,
    $4
)
ON CONFLICT (address)
DO UPDATE SET address = EXCLUDED.address
RETURNING *;

-- name: GetMailboxId :one
SELECT id from mailboxes where address=$1;