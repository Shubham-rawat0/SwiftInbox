-- name: CreateEmailAddress :one
INSERT INTO mailboxes(
    id , address , expires_at
) 
VALUES ($1,$2,$3) 
ON CONFLICT(address)
DO UPDATE SET expires_at= EXCLUDED.expires_at
Returning address, created_at, expires_at;

-- name: UpsertMailbox :one

INSERT INTO mailboxes (
    id,
    address,
    expires_at
)
VALUES (
    $1,
    $2,
    $3
)
ON CONFLICT (address)
DO UPDATE SET address = EXCLUDED.address
RETURNING *;

-- name: GetMailboxId :one
SELECT id from mailboxes where address=$1;