package queue

import (
	"context"
	"fmt"

	"github.com/Shubham-rawat0/temp-mail/SwiftInbox/backend/internal/repository/postgres"
	"github.com/google/uuid"
)

func PublishForMailbox(
	ctx context.Context,
	queries *postgres.Queries,
	publisher Publisher,
	mailboxID uuid.UUID,
	eventName string,
	messageID uuid.UUID,
) error {
	webhooks, err := queries.GetWebhooksByMailboxID(ctx, mailboxID)
	if err != nil {
		return fmt.Errorf("find webhooks for mailbox %s: %w", mailboxID, err)
	}

	return PublishForWebhooks(ctx, publisher, webhooks, mailboxID, eventName, messageID)
}

func PublishForWebhooks(
	ctx context.Context,
	publisher Publisher,
	webhooks []postgres.GetWebhooksByMailboxIDRow,
	mailboxID uuid.UUID,
	eventName string,
	messageID uuid.UUID,
) error {

	var lastErr error
	for _, webhook := range webhooks {
		if !webhook.IsActive || !subscribesTo(webhook.Events, eventName) {
			continue
		}

		event := WebhookEvent{
			DeveloperID:     webhook.DeveloperID.String(),
			WebhookID:       webhook.ID.String(),
			MailboxID:       mailboxID.String(),
			URL:             webhook.Url,
			Event:           eventName,
			SecretEncrypted: webhook.SecretEncrypted,
		}
		if messageID != uuid.Nil {
			event.MessageID = messageID.String()
		}

		if err := publisher.Publish(ctx, event); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

func subscribesTo(events []string, eventName string) bool {
	for _, subscribedEvent := range events {
		if subscribedEvent == eventName {
			return true
		}
	}
	return false
}
