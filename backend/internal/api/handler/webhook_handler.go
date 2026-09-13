package handler

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/utils"
	"github.com/google/uuid"
)

type WebhookHandler struct {
	queries *postgres.Queries
}

func NewWebhookHandler(q *postgres.Queries) *WebhookHandler {
	return &WebhookHandler{
		queries: q,
	}
}

type WebhookReqBody struct {
	Url        string      `json:"url"`
	Events     []string    `json:"events"`
	MailboxIDs []uuid.UUID `json:"mailbox_ids"`
}

type WebhookEventsBody struct {
	Events []string `json:"events"`
}

type WebhookMailboxBody struct {
	MailboxID uuid.UUID `json:"mailbox_id"`
}

type WebhookResBody struct {
	ID          uuid.UUID   `json:"id"`
	DeveloperID uuid.UUID   `json:"developer_id"`
	Url         string      `json:"url"`
	IsActive    bool        `json:"is_active"`
	Events      []string    `json:"events"`
	Secret      string      `json:"secret"`
	MailboxIDs  []uuid.UUID `json:"mailbox_ids"`
}

type webhookResponse struct {
	ID          uuid.UUID `json:"id"`
	DeveloperID uuid.UUID `json:"developer_id"`
	Url         string    `json:"url"`
	IsActive    bool      `json:"is_active"`
	Events      []string  `json:"events"`
}

func (a *WebhookHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	reqBody := WebhookReqBody{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		WriteError(w, http.StatusBadRequest, errors.New("provide complete req body"))
		return
	}

	err := utils.ValidateEvents(reqBody.Events)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	if len(reqBody.MailboxIDs) == 0 {
		WriteError(w, http.StatusBadRequest, errors.New("mailbox_ids are required"))
		return
	}

	id := uuid.New()
	developerId, ok := utils.DeveloperIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusBadRequest, errors.New("No developer id/api key"))
		return
	}

	secret, err := utils.GenerateSecret()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, errors.New("error generating secret"))
		return
	}

	encryptionKey, err := webhookEncryptionKey()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

	encryptSecret, err := utils.Encrypt(secret, encryptionKey)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, errors.New("error encrypting secret"))
		return
	}

	data, err := a.queries.CreateWebhook(r.Context(), postgres.CreateWebhookParams{
		ID:              id,
		DeveloperID:     developerId,
		Url:             reqBody.Url,
		SecretEncrypted: encryptSecret,
		Events:          reqBody.Events,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(w, http.StatusNotFound, err)
			return
		}

		WriteError(w, http.StatusInternalServerError, errors.New("failed to create webhook"))
		return
	}

	linkedMailboxIDs := make([]uuid.UUID, 0, len(reqBody.MailboxIDs))
	seenMailboxIDs := make(map[uuid.UUID]struct{}, len(reqBody.MailboxIDs))
	for _, mailboxID := range reqBody.MailboxIDs {
		if _, seen := seenMailboxIDs[mailboxID]; seen {
			continue
		}
		seenMailboxIDs[mailboxID] = struct{}{}

		linkedMailboxID, linkErr := a.queries.LinkWebhookMailbox(r.Context(), postgres.LinkWebhookMailboxParams{
			WebhookID:   id,
			ID:          mailboxID,
			DeveloperID: developerId,
		})
		if linkErr != nil {
			_, _ = a.queries.DeleteWebhook(r.Context(), postgres.DeleteWebhookParams{ID: id, DeveloperID: developerId})
			if errors.Is(linkErr, sql.ErrNoRows) {
				WriteError(w, http.StatusBadRequest, errors.New("one or more mailboxes do not belong to the developer"))
				return
			}
			WriteError(w, http.StatusInternalServerError, errors.New("failed to link mailbox to webhook"))
			return
		}
		linkedMailboxIDs = append(linkedMailboxIDs, linkedMailboxID)
	}

	res := WebhookResBody{ID: data.ID,
		DeveloperID: data.DeveloperID,
		Url:         data.Url,
		IsActive:    data.IsActive,
		Events:      data.Events,
		Secret:      secret,
		MailboxIDs:  linkedMailboxIDs,
	}
	WriteJSON(w, http.StatusOK, res)
}

func (a *WebhookHandler) AddMailbox(w http.ResponseWriter, r *http.Request) {
	webhookID, developerID, ok := a.webhookContext(w, r)
	if !ok {
		return
	}

	var request WebhookMailboxBody
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.MailboxID == uuid.Nil {
		WriteError(w, http.StatusBadRequest, errors.New("mailbox_id is required"))
		return
	}

	mailboxID, err := a.queries.LinkWebhookMailbox(r.Context(), postgres.LinkWebhookMailboxParams{
		WebhookID:   webhookID,
		ID:          request.MailboxID,
		DeveloperID: developerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(w, http.StatusNotFound, errors.New("webhook or mailbox not found"))
			return
		}
		WriteError(w, http.StatusInternalServerError, errors.New("failed to add mailbox to webhook"))
		return
	}

	WriteJSON(w, http.StatusCreated, map[string]string{
		"message":    "mailbox added to webhook",
		"mailbox_id": mailboxID.String(),
	})
}

func (a *WebhookHandler) RemoveMailbox(w http.ResponseWriter, r *http.Request) {
	webhookID, developerID, ok := a.webhookContext(w, r)
	if !ok {
		return
	}

	mailboxID, err := uuid.Parse(r.PathValue("mailboxID"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, errors.New("invalid mailbox id"))
		return
	}

	removedMailboxID, err := a.queries.RemoveWebhookMailbox(r.Context(), postgres.RemoveWebhookMailboxParams{
		ID:          webhookID,
		MailboxID:   mailboxID,
		DeveloperID: developerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(w, http.StatusNotFound, errors.New("webhook mailbox link not found"))
			return
		}
		WriteError(w, http.StatusInternalServerError, errors.New("failed to remove mailbox from webhook"))
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message":    "mailbox removed from webhook",
		"mailbox_id": removedMailboxID.String(),
	})
}

func (a *WebhookHandler) GetWebhook(w http.ResponseWriter, r *http.Request) {
	webhookID, developerID, ok := a.webhookContext(w, r)
	if !ok {
		return
	}

	data, err := a.queries.GetWebhook(r.Context(), postgres.GetWebhookParams{
		ID: webhookID, DeveloperID: developerID,
	})
	if err != nil {
		a.writeWebhookQueryError(w, err, "failed to fetch webhook")
		return
	}

	WriteJSON(w, http.StatusOK, webhookResponse{
		ID: data.ID, DeveloperID: data.DeveloperID, Url: data.Url,
		IsActive: data.IsActive, Events: data.Events,
	})
}

func (a *WebhookHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	developerID, ok := utils.DeveloperIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, errors.New("api key required"))
		return
	}

	data, err := a.queries.ListWebhooks(r.Context(), developerID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, errors.New("failed to fetch webhooks"))
		return
	}

	webhooks := make([]webhookResponse, 0, len(data))
	for _, webhook := range data {
		webhooks = append(webhooks, webhookResponse{
			ID: webhook.ID, DeveloperID: webhook.DeveloperID, Url: webhook.Url,
			IsActive: webhook.IsActive, Events: webhook.Events,
		})
	}

	WriteJSON(w, http.StatusOK, webhooks)
}

func (a *WebhookHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	webhookID, developerID, ok := a.webhookContext(w, r)
	if !ok {
		return
	}

	data, err := a.queries.DeleteWebhook(r.Context(), postgres.DeleteWebhookParams{
		ID: webhookID, DeveloperID: developerID,
	})
	if err != nil {
		a.writeWebhookQueryError(w, err, "failed to delete webhook")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "webhook deleted",
		"id":      data.String(),
	})
}

func (a *WebhookHandler) AddEvents(w http.ResponseWriter, r *http.Request) {
	a.updateEvents(w, r, true)
}

func (a *WebhookHandler) RemoveEvents(w http.ResponseWriter, r *http.Request) {
	a.updateEvents(w, r, false)
}

func (a *WebhookHandler) updateEvents(w http.ResponseWriter, r *http.Request, add bool) {
	webhookID, developerID, ok := a.webhookContext(w, r)
	if !ok {
		return
	}

	request := WebhookEventsBody{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteError(w, http.StatusBadRequest, errors.New("invalid event body"))
		return
	}
	if len(request.Events) == 0 {
		WriteError(w, http.StatusBadRequest, errors.New("events are required"))
		return
	}
	if err := utils.ValidateEvents(request.Events); err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	if add {
		data, err := a.queries.InsertEvent(r.Context(), postgres.InsertEventParams{
			ID: webhookID, DeveloperID: developerID, Events: request.Events,
		})
		if err != nil {
			a.writeWebhookQueryError(w, err, "failed to update webhook events")
			return
		}
		WriteJSON(w, http.StatusOK, webhookResponse{
			ID: data.ID, DeveloperID: data.DeveloperID, Url: data.Url,
			IsActive: data.IsActive, Events: data.Events,
		})
		return
	} else {
		data, err := a.queries.DeleteEvents(r.Context(), postgres.DeleteEventsParams{
			ID: webhookID, DeveloperID: developerID, Column3: request.Events,
		})
		if err != nil {
			a.writeWebhookQueryError(w, err, "failed to update webhook events")
			return
		}
		WriteJSON(w, http.StatusOK, webhookResponse{
			ID: data.ID, DeveloperID: data.DeveloperID, Url: data.Url,
			IsActive: data.IsActive, Events: data.Events,
		})
	}
}

func (a *WebhookHandler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	webhookID, developerID, ok := a.webhookContext(w, r)
	if !ok {
		return
	}

	data, err := a.queries.GetWebhook(r.Context(), postgres.GetWebhookParams{
		ID: webhookID, DeveloperID: developerID,
	})
	if err != nil {
		a.writeWebhookQueryError(w, err, "failed to fetch webhook")
		return
	}
	if data.ID != webhookID {
		WriteError(w, http.StatusNotFound, errors.New("webhook not found"))
		return
	}
	if !data.IsActive {
		WriteError(w, http.StatusConflict, errors.New("webhook is inactive"))
		return
	}

	encryptionKey, err := webhookEncryptionKey()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err)
		return
	}
	secret, err := utils.Decrypt(data.SecretEncrypted, encryptionKey)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, errors.New("failed to decrypt webhook secret"))
		return
	}

	payload, err := json.Marshal(map[string]any{
		"event": "webhook.test",
		"data":  map[string]string{"webhook_id": data.ID.String()},
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, errors.New("failed to encode test payload"))
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, data.Url, bytes.NewReader(payload))
	if err != nil {
		WriteError(w, http.StatusBadRequest, errors.New("invalid webhook url"))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Event", "webhook.test")
	req.Header.Set("X-Webhook-Signature", utils.Sign(payload, secret)) //signature for payload

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		WriteError(w, http.StatusBadGateway, errors.New("webhook request failed"))
		return
	}
	defer resp.Body.Close()

	WriteJSON(w, http.StatusOK, map[string]any{
		"success":    resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices,
		"statusCode": resp.StatusCode,
	})
}

func (a *ApiHandler) GetWebhookDeadLetters(w http.ResponseWriter, r *http.Request) {
	developerID, ok := utils.DeveloperIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, errors.New("api key required"))
		return
	}

	data, err := a.queries.ListWebhookDeadLetters(r.Context(), developerID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, errors.New("failed to fetch webhook dead letters"))
		return
	}

	deadLetters := make([]map[string]any, 0, len(data))
	for _, item := range data {
		deadLetters = append(deadLetters, map[string]any{
			"id":         item.ID,
			"webhook_id": item.WebhookID,
			"mailbox_id": nullableUUIDString(item.MailboxID),
			"message_id": nullableUUIDString(item.MessageID),
			"event":      item.Event,
			"url":        item.Url,
			"reason":     item.Reason,
			"attempts":   item.Attempts,
			"created_at": item.CreatedAt,
		})
	}

	WriteJSON(w, http.StatusOK, deadLetters)
}

func nullableUUIDString(value uuid.NullUUID) any {
	if !value.Valid {
		return nil
	}
	return value.UUID
}


func webhookEncryptionKey() ([]byte, error) {
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

func (a *WebhookHandler) webhookContext(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	webhookID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, errors.New("invalid webhook id"))
		return uuid.Nil, uuid.Nil, false
	}

	developerID, ok := utils.DeveloperIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, errors.New("api key required"))
		return uuid.Nil, uuid.Nil, false
	}

	return webhookID, developerID, true
}

func (a *WebhookHandler) writeWebhookQueryError(w http.ResponseWriter, err error, message string) {
	if errors.Is(err, sql.ErrNoRows) {
		WriteError(w, http.StatusNotFound, errors.New("webhook not found"))
		return
	}
	WriteError(w, http.StatusInternalServerError, errors.New(message))
}
