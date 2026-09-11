-- name: IncrementApiKeyUsage :one
INSERT INTO api_key_usage (
    api_key_id,
    period,
    api_requests,
    mailbox_requests,
    message_requests
)
VALUES (
    @api_key_id,
    DATE_TRUNC('month', CURRENT_DATE)::DATE,
    @api_requests,
    CASE WHEN @category = 'mailbox' THEN @api_requests ELSE 0 END,
    CASE WHEN @category = 'message' THEN @api_requests ELSE 0 END
)
ON CONFLICT (api_key_id, period)
DO UPDATE SET
    api_requests = api_key_usage.api_requests + EXCLUDED.api_requests,
    mailbox_requests = api_key_usage.mailbox_requests + EXCLUDED.mailbox_requests,
    message_requests = api_key_usage.message_requests + EXCLUDED.message_requests
RETURNING api_key_id, period, api_requests, mailbox_requests, message_requests;