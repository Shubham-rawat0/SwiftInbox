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
    COALESCE(du.api_requests, 0) AS api_requests,
    COALESCE(du.mailbox_requests, 0) AS mailbox_requests,
    COALESCE(du.messages_requests, 0) AS messages_requests
FROM developer d
LEFT JOIN developer_usage du
    ON d.id = du.developer_id
   AND du.period = DATE_TRUNC('month', CURRENT_DATE)::DATE
WHERE d.id = $1;