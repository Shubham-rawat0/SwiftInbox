package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/utils"
	"github.com/google/uuid"
)

var (
	ErrWebhookNotFound            = errors.New("webhook not found")
	ErrWebhookInactive            = errors.New("webhook is inactive")
	ErrMailboxNotOwned            = errors.New("one or more mailboxes do not belong to the developer")
	ErrWebhookMailboxNotFound     = errors.New("webhook or mailbox not found")
	ErrWebhookMailboxLinkNotFound = errors.New("webhook mailbox link not found")
	ErrEventsRequired             = errors.New("events are required")
	ErrMailboxIDsRequired         = errors.New("mailbox_ids are required")
	ErrInvalidWebhookEvent        = errors.New("invalid webhook event")
	ErrMailboxAlreadyLinked       = errors.New("mailbox already linked to webhook")
)

type WebhookResult struct {
	ID          uuid.UUID   `json:"id"`
	DeveloperID uuid.UUID   `json:"developer_id"`
	Url         string      `json:"url"`
	IsActive    bool        `json:"is_active"`
	Events      []string    `json:"events"`
	Secret      string      `json:"secret,omitempty"`
	MailboxIDs  []uuid.UUID `json:"mailbox_ids,omitempty"`
}

type DeadLetterResult struct {
	ID        uuid.UUID     `json:"id"`
	WebhookID uuid.UUID     `json:"webhook_id"`
	MailboxID uuid.NullUUID `json:"mailbox_id"`
	MessageID uuid.NullUUID `json:"message_id"`
	Event     string        `json:"event"`
	Url       string        `json:"url"`
	Reason    string        `json:"reason"`
	Attempts  int32         `json:"attempts"`
	CreatedAt time.Time     `json:"created_at"`
}

type WebhookService struct {
	queries *postgres.Queries
}

func NewWebhookService(q *postgres.Queries) *WebhookService {
	return &WebhookService{
		queries: q,
	}
}

func WebhookEncryptionKey() ([]byte, error) {
	configuredKey := os.Getenv("WEBHOOK_ENCRYPTION_KEY")
	if len(configuredKey) == utils.KeyBytes {
		return []byte(configuredKey), nil
	}

	decodedKey, err := base64.RawStdEncoding.DecodeString(configuredKey)
	if err != nil {
		decodedKey, err = base64.StdEncoding.DecodeString(configuredKey)
	}
	if err != nil || len(decodedKey) != utils.KeyBytes {
		return nil, errors.New("WEBHOOK_ENCRYPTION_KEY must be a 32-byte key or base64-encoded 32-byte key")
	}

	return decodedKey, nil
}

func (s *WebhookService) CreateWebhook(ctx context.Context, devID uuid.UUID, url string, events []string, mailboxIDs []uuid.UUID) (*WebhookResult, error) {
	if err := ValidateEvents(events); err != nil {
		return nil, err
	}
	if len(mailboxIDs) == 0 {
		return nil, ErrMailboxIDsRequired
	}

	secret, err := utils.GenerateSecret()
	if err != nil {
		return nil, errors.New("error generating secret")
	}

	encKey, err := WebhookEncryptionKey()
	if err != nil {
		return nil, err
	}

	encryptSecret, err := utils.Encrypt(secret, encKey)
	if err != nil {
		return nil, errors.New("error encrypting secret")
	}

	id := uuid.New()
	data, err := s.queries.CreateWebhook(ctx, postgres.CreateWebhookParams{
		ID:              id,
		DeveloperID:     devID,
		Url:             url,
		SecretEncrypted: encryptSecret,
		Events:          events,
	})
	if err != nil {
		return nil, err
	}

	linkedMailboxIDs := make([]uuid.UUID, 0, len(mailboxIDs))
	seenMailboxIDs := make(map[uuid.UUID]struct{}, len(mailboxIDs))
	for _, mailboxID := range mailboxIDs {
		if _, seen := seenMailboxIDs[mailboxID]; seen {
			continue
		}
		seenMailboxIDs[mailboxID] = struct{}{}

		linkedMailboxID, linkErr := s.queries.LinkWebhookMailbox(ctx, postgres.LinkWebhookMailboxParams{
			WebhookID:   id,
			ID:          mailboxID,
			DeveloperID: devID,
		})
		if linkErr != nil {
			_, _ = s.queries.DeleteWebhook(ctx, postgres.DeleteWebhookParams{ID: id, DeveloperID: devID})
			if errors.Is(linkErr, sql.ErrNoRows) {
				return nil, ErrMailboxNotOwned
			}
			return nil, errors.New("failed to link mailbox to webhook")
		}
		linkedMailboxIDs = append(linkedMailboxIDs, linkedMailboxID)
	}

	return &WebhookResult{
		ID:          data.ID,
		DeveloperID: data.DeveloperID,
		Url:         data.Url,
		IsActive:    data.IsActive,
		Events:      data.Events,
		Secret:      secret,
		MailboxIDs:  linkedMailboxIDs,
	}, nil
}

func (s *WebhookService) ListWebhooks(ctx context.Context, devID uuid.UUID) ([]WebhookResult, error) {
	data, err := s.queries.ListWebhooks(ctx, devID)
	if err != nil {
		return nil, err
	}

	webhooks := make([]WebhookResult, 0, len(data))
	for _, w := range data {
		webhooks = append(webhooks, WebhookResult{
			ID:          w.ID,
			DeveloperID: w.DeveloperID,
			Url:         w.Url,
			IsActive:    w.IsActive,
			Events:      w.Events,
		})
	}
	return webhooks, nil
}

func (s *WebhookService) GetWebhook(ctx context.Context, devID uuid.UUID, webhookID uuid.UUID) (*WebhookResult, error) {
	data, err := s.queries.GetWebhook(ctx, postgres.GetWebhookParams{
		ID:          webhookID,
		DeveloperID: devID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWebhookNotFound
		}
		return nil, err
	}

	return &WebhookResult{
		ID:          data.ID,
		DeveloperID: data.DeveloperID,
		Url:         data.Url,
		IsActive:    data.IsActive,
		Events:      data.Events,
	}, nil
}

func (s *WebhookService) DeleteWebhook(ctx context.Context, devID uuid.UUID, webhookID uuid.UUID) error {
	_, err := s.queries.DeleteWebhook(ctx, postgres.DeleteWebhookParams{
		ID:          webhookID,
		DeveloperID: devID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrWebhookNotFound
		}
		return err
	}
	return nil
}

func (s *WebhookService) AddMailbox(ctx context.Context, devID uuid.UUID, webhookID uuid.UUID, mailboxID uuid.UUID) (uuid.UUID, error) {
	resID, err := s.queries.LinkWebhookMailbox(ctx, postgres.LinkWebhookMailboxParams{
		WebhookID:   webhookID,
		ID:          mailboxID,
		DeveloperID: devID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if _, lookupErr := s.queries.GetWebhookMailbox(ctx, postgres.GetWebhookMailboxParams{
				WebhookID:   webhookID,
				MailboxID:   mailboxID,
				DeveloperID: devID,
			}); lookupErr == nil {
				return uuid.Nil, ErrMailboxAlreadyLinked
			}
			return uuid.Nil, ErrWebhookMailboxNotFound
		}
		return uuid.Nil, err
	}
	return resID, nil
}

func (s *WebhookService) RemoveMailbox(ctx context.Context, devID uuid.UUID, webhookID uuid.UUID, mailboxID uuid.UUID) (uuid.UUID, error) {
	resID, err := s.queries.RemoveWebhookMailbox(ctx, postgres.RemoveWebhookMailboxParams{
		ID:          webhookID,
		MailboxID:   mailboxID,
		DeveloperID: devID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, ErrWebhookMailboxLinkNotFound
		}
		return uuid.Nil, err
	}
	return resID, nil
}

func (s *WebhookService) AddEvents(ctx context.Context, devID uuid.UUID, webhookID uuid.UUID, events []string) (*WebhookResult, error) {
	if len(events) == 0 {
		return nil, ErrEventsRequired
	}
	if err := ValidateEvents(events); err != nil {
		return nil, err
	}

	data, err := s.queries.InsertEvent(ctx, postgres.InsertEventParams{
		ID:          webhookID,
		DeveloperID: devID,
		Events:      events,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWebhookNotFound
		}
		return nil, err
	}

	return &WebhookResult{
		ID:          data.ID,
		DeveloperID: data.DeveloperID,
		Url:         data.Url,
		IsActive:    data.IsActive,
		Events:      data.Events,
	}, nil
}

func (s *WebhookService) RemoveEvents(ctx context.Context, devID uuid.UUID, webhookID uuid.UUID, events []string) (*WebhookResult, error) {
	if len(events) == 0 {
		return nil, ErrEventsRequired
	}
	if err := ValidateEvents(events); err != nil {
		return nil, err
	}

	data, err := s.queries.DeleteEvents(ctx, postgres.DeleteEventsParams{
		ID:          webhookID,
		DeveloperID: devID,
		Column3:     events,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWebhookNotFound
		}
		return nil, err
	}

	return &WebhookResult{
		ID:          data.ID,
		DeveloperID: data.DeveloperID,
		Url:         data.Url,
		IsActive:    data.IsActive,
		Events:      data.Events,
	}, nil
}

func (s *WebhookService) TestWebhook(ctx context.Context, devID uuid.UUID, webhookID uuid.UUID) (int, bool, error) {
	data, err := s.queries.GetWebhook(ctx, postgres.GetWebhookParams{
		ID:          webhookID,
		DeveloperID: devID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, ErrWebhookNotFound
		}
		return 0, false, err
	}
	if !data.IsActive {
		return 0, false, ErrWebhookInactive
	}

	encKey, err := WebhookEncryptionKey()
	if err != nil {
		return 0, false, err
	}
	secret, err := utils.Decrypt(data.SecretEncrypted, encKey)
	if err != nil {
		return 0, false, errors.New("failed to decrypt webhook secret")
	}

	payload, err := json.Marshal(map[string]any{
		"event": "webhook.test",
		"data":  map[string]string{"webhook_id": data.ID.String()},
	})
	if err != nil {
		return 0, false, errors.New("failed to encode test payload")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, data.Url, bytes.NewReader(payload))
	if err != nil {
		return 0, false, errors.New("invalid webhook url")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Event", "webhook.test")
	req.Header.Set("X-Webhook-Signature", utils.Sign(payload, secret))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, false, errors.New("webhook request failed")
	}
	defer resp.Body.Close()

	success := resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices
	return resp.StatusCode, success, nil
}

func (s *WebhookService) ListDeadLetters(ctx context.Context, devID uuid.UUID) ([]DeadLetterResult, error) {
	data, err := s.queries.ListWebhookDeadLetters(ctx, devID)
	if err != nil {
		return nil, err
	}

	results := make([]DeadLetterResult, 0, len(data))
	for _, item := range data {
		results = append(results, DeadLetterResult{
			ID:        item.ID,
			WebhookID: item.WebhookID,
			MailboxID: item.MailboxID,
			MessageID: item.MessageID,
			Event:     item.Event,
			Url:       item.Url,
			Reason:    item.Reason,
			Attempts:  item.Attempts,
			CreatedAt: item.CreatedAt,
		})
	}
	return results, nil
}
