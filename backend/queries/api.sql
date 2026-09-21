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

-- name: UpdateApiKeyLastUsed :one
UPDATE apikeys
SET last_used_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, last_used_at;

-- name: RevokeApiKey :one
UPDATE apikeys
SET revoked_at = COALESCE(revoked_at, CURRENT_TIMESTAMP)
WHERE id = $1
RETURNING id, last_used_at, revoked_at;

-- name: RevokeDeveloperApiKey :one
UPDATE apikeys
SET revoked_at = COALESCE(revoked_at, CURRENT_TIMESTAMP)
WHERE id = $1 AND developer_id = $2
RETURNING id, last_used_at, revoked_at;


-- name: GetUserApiKeys :many

SELECT
    a.id,
    a.name,
    a.last_used_at,
    a.revoked_at,
    a.created_at,
    COALESCE(b.api_requests, 0) AS api_requests,
    COALESCE(b.mailbox_requests, 0) AS mailbox_requests,
    COALESCE(b.message_requests, 0) AS message_requests
FROM apikeys a
LEFT JOIN api_key_usage b
    ON b.api_key_id = a.id
   AND b.period = DATE_TRUNC('month', CURRENT_DATE)::DATE
WHERE a.developer_id = $1;

