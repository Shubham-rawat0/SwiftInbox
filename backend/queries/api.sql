-- name: CreateApiKey :one
INSERT INTO apikeys (
    id ,
    developer_id,
    key_hash,
    last_used_at,
    revoked_at 
) VALUES ($1,$2,$3,$4,$5)
RETURNING id, developer_id, key_hash;

-- name: GetApiKey :one
SELECT id, developer_id , last_used_at ,revoked_at from apikeys where key_hash =$1;

-- name: RevokeApiKey :one
DELETE FROM apikeys WHERE id=$1 RETURNING id,last_used_at, revoked_at;