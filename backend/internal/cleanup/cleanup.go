package cleanup

import (
	"context"
	"database/sql"
	"log"
	"time"

	queue "github.com/Shubham-rawat0/temp-mail/SwiftInbox/backend/internal/events"
	"github.com/Shubham-rawat0/temp-mail/SwiftInbox/backend/internal/repository/postgres"
	"github.com/google/uuid"
)

const (
	revokedApiKeyRetentionFallback = 7 * 24 * time.Hour
)

func CleanUpExpired(q *postgres.Queries, publisher queue.Publisher) (int32, error) {
	ctx := context.Background()
	now := time.Now()

	retention := getDurationEnv(
		"REVOKED_API_KEY_RETENTION_MS",
		revokedApiKeyRetentionFallback,
	)

	deletedRevokedApiKeys, err := q.DeleteRevokedApiKeys(
		ctx,
		sql.NullTime{Time: now.Add(-retention), Valid: true},
	)
	if err != nil {
		log.Println("Revoked API key cleanup error:", err.Error())
		return 0, err
	}

	deletedMessages, err := q.DeleteExpiredMessages(ctx, now)
	if err != nil {
		log.Println("Message cleanup error:", err.Error())
		return 0, err
	}
	for _, message := range deletedMessages {
		if err := queue.PublishForMailbox(ctx, q, publisher, message.MailboxID, "email.deleted", message.ID); err != nil {
			log.Printf("[WEBHOOK ERROR] event=email.deleted mailbox=%s message=%s error=%v", message.MailboxID, message.ID, err)
		}
	}

	expiredMailboxIDs, err := q.GetExpiredMailboxIDs(ctx, now)
	if err != nil {
		log.Println("Mailbox lookup error:", err.Error())
		return int32(len(deletedMessages)), err
	}

	webhooksByMailbox := make(map[uuid.UUID][]postgres.GetWebhooksByMailboxIDRow, len(expiredMailboxIDs))
	for _, mailboxID := range expiredMailboxIDs {
		webhooks, err := q.GetWebhooksByMailboxID(ctx, mailboxID)
		if err != nil {
			log.Printf("Mailbox webhook lookup error mailbox=%s: %v", mailboxID, err)
			return int32(len(deletedMessages)), err
		}
		webhooksByMailbox[mailboxID] = webhooks
	}

	deletedMailboxIDs, err := q.DeleteExpiredMailbox(ctx, now)
	if err != nil {
		log.Println("Mailbox cleanup error:", err.Error())
		return int32(len(deletedMessages)), err
	}
	for _, mailboxID := range deletedMailboxIDs {
		if err := queue.PublishForWebhooks(ctx, publisher, webhooksByMailbox[mailboxID], mailboxID, "mailbox.expired", uuid.Nil); err != nil {
			log.Printf("[WEBHOOK ERROR] event=mailbox.expired mailbox=%s error=%v", mailboxID, err)
		}
	}

	count := int32(len(deletedMessages) + len(deletedMailboxIDs) + len(deletedRevokedApiKeys))
	if count == 0 {
		log.Println("nothing to delete")
	} else {
		log.Printf("deleted %d messages, %d mailboxes and %d revoked api keys", len(deletedMessages), len(deletedMailboxIDs), len(deletedRevokedApiKeys))
	}
	return count, nil
}
