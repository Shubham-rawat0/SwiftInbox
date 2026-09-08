-- name: IncrementUsage :one
INSERT INTO developer_usage (
    developer_id,
    period,
    api_requests
)
VALUES (
    $1,
    DATE_TRUNC('month', CURRENT_DATE)::DATE,
    1
)
ON CONFLICT (developer_id, period)
DO UPDATE SET
    api_requests = developer_usage.api_requests + 1
RETURNING *;

