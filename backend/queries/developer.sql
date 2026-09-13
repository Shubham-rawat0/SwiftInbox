-- name: CreateDeveloper :one
INSERT INTO developer (
    id ,
    name,
    email,
    password_hash
) VALUES ($1,$2,$3,$4)
RETURNING id, name,email;

-- name: GetDeveloperAuthByEmail :one
SELECT id, name, email
    , password_hash
FROM developer
WHERE email = $1;

-- name: GetDeveloperPasswordHash :one
SELECT password_hash
FROM developer
WHERE id = $1;

-- name: GetDeveloper :one

SELECT
    d.name,
    d.email,
    d.api_quota,
    d.mailbox_quota,
    d.message_quota,
    d.created_at,
    COALESCE(du.api_requests, 0) AS api_requests,
    COALESCE(du.mailbox_requests, 0) AS mailbox_requests,
    COALESCE(du.messages_requests, 0) AS messages_requests
FROM developer d
LEFT JOIN developer_usage du
    ON d.id = du.developer_id
   AND du.period = DATE_TRUNC('month', CURRENT_DATE)::DATE
WHERE d.id = $1;