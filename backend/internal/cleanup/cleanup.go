package cleanup

import (
	"context"
	"log"
	"time"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
)

func CleanUpExpired(q *postgres.Queries) (int32, error) {
	ctx := context.Background()
	now := time.Now()

	messageCount, err := q.DeleteExpiredMessages(ctx, now)
	if err != nil {
		log.Println("Message cleanup error:", err.Error())
		return 0, err
	}

	mailboxCount, err := q.DeleteExpiredMailbox(ctx, now)
	if err != nil {
		log.Println("Mailbox cleanup error:", err.Error())
		return messageCount, err
	}

	count := messageCount + mailboxCount
	if count == 0 {
		log.Println("nothing to delete")
	} else {
		log.Printf("deleted %d messages and %d mailboxes", messageCount, mailboxCount)
	}
	return count, nil
}
