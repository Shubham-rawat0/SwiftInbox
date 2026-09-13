package cleanup

import (
	"context"
	"log"
	"time"

	queue "github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/events"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/google/uuid"
)

func CleanUpExpired(q *postgres.Queries, publisher queue.Publisher) (int32, error) {
	ctx := context.Background()
	now := time.Now()

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

	count := int32(len(deletedMessages) + len(deletedMailboxIDs))
	if count == 0 {
		log.Println("nothing to delete")
	} else {
		log.Printf("deleted %d messages and %d mailboxes", len(deletedMessages), len(deletedMailboxIDs))
	}
	return count, nil
}
