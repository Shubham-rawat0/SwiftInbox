-- name: CreateDeveloper :one
INSERT INTO developer (
    id ,
    name,
    email
) VALUES ($1,$2,$3)
RETURNING id, name,email;

-- name: GetDeveloper :one
SELECT
    d.name,
    d.email,
    d.api_quota,
    d.mailbox_quota,
    d.message_quota,
    d.created_at,
    du.api_requests,
    du.mailbox_requests,
    du.messages_requests
FROM developer d
JOIN developer_usage du
    ON d.id = du.developer_id
WHERE d.id = $1
  AND du.period = DATE_TRUNC('month', CURRENT_DATE)::DATE;