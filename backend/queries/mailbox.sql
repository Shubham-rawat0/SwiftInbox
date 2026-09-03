-- name: CreateEmailAddress :one
INSERT INTO mailboxes(
    id , address , expires_at
) 
VALUES ($1,$2,$3) 
ON CONFLICT(address)
DO UPDATE SET expires_at= EXCLUDED.expires_at
Returning address, created_at, expires_at;