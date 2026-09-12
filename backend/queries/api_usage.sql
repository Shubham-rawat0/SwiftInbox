-- name: IncrementUsage :one
INSERT INTO developer_usage (developer_id, period, api_requests, mailbox_requests, messages_requests)
SELECT @developer_id, DATE_TRUNC('month', CURRENT_DATE)::DATE, @api_requests,
             CASE WHEN @category = 'mailbox' THEN @api_requests ELSE 0 END,
             CASE WHEN @category = 'message' THEN @api_requests ELSE 0 END
FROM developer
WHERE id = @developer_id
    AND @api_requests <= api_quota
  AND (
            (@category = 'mailbox' AND @api_requests <= mailbox_quota)
            OR (@category = 'message' AND @api_requests <= message_quota)
            OR (@category = 'api')
  )
ON CONFLICT (developer_id, period)
DO UPDATE SET
    api_requests = developer_usage.api_requests + EXCLUDED.api_requests,
    mailbox_requests = developer_usage.mailbox_requests + EXCLUDED.mailbox_requests,
    messages_requests = developer_usage.messages_requests + EXCLUDED.messages_requests
WHERE developer_usage.api_requests + EXCLUDED.api_requests <= (
          SELECT api_quota FROM developer WHERE id = @developer_id
      )
  AND (
      (@category = 'mailbox' AND developer_usage.mailbox_requests + EXCLUDED.mailbox_requests <= (
          SELECT mailbox_quota FROM developer WHERE id = @developer_id
      ))
      OR (@category = 'message' AND developer_usage.messages_requests + EXCLUDED.messages_requests <= (
          SELECT message_quota FROM developer WHERE id = @developer_id
      ))
      OR (@category = 'api')
  )
RETURNING developer_id, period, api_requests, mailbox_requests, messages_requests;
