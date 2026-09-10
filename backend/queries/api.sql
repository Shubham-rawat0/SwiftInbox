-- name: CreateApiKey :one
INSERT INTO apikeys (
    id ,
    developer_id,
    name,
    key_hash,
    last_used_at,
    revoked_at 
) VALUES ($1,$2,$3,$4,$5,$6)
RETURNING id, developer_id, name, key_hash;

-- name: GetApiKey :one
SELECT id, developer_id, name, last_used_at, revoked_at from apikeys where key_hash =$1;

-- name: RevokeApiKey :one
UPDATE apikeys
SET revoked_at = COALESCE(revoked_at, CURRENT_TIMESTAMP)
WHERE id = $1
RETURNING id, last_used_at, revoked_at;

-- name: GetUserApiKeys :many
SELECT id,name,last_used_at,revoked_at,created_at from apikeys where developer_id=$1;