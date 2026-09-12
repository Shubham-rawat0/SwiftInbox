package handler

import (
	"bytes"
	"database/sql"
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
	Url    string   `json:"url"`
	Events []string `json:"events"`
}

type WebhookEventsBody struct {
	Events []string `json:"events"`
}

type WebhookResBody struct {
	ID          uuid.UUID `json:"id"`
	DeveloperID uuid.UUID `json:"developer_id"`
	Url         string    `json:"url"`
	IsActive    bool      `json:"is_active"`
	Events      []string  `json:"events"`
	Secret      string    `json:"secret"`
}

type webhookResponse struct {
	ID          uuid.UUID `json:"id"`
	DeveloperID uuid.UUID `json:"developer_id"`
	Url         string    `json:"url"`
	IsActive    bool      `json:"is_active"`
	Events      []string  `json:"events"`
}

var webhook_enc_key = os.Getenv("WEBHOOK_ENCRYPTION_KEY")

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

	encryptSecret, err := utils.Encrypt(secret, []byte(webhook_enc_key))
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
	res := WebhookResBody{ID: data.ID,
		DeveloperID: data.DeveloperID,
		Url:         data.Url,
		IsActive:    data.IsActive,
		Events:      data.Events,
		Secret:      secret,
	}
	WriteJSON(w, http.StatusOK, res)
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

	secret, err := utils.Decrypt(data.SecretEncrypted, []byte(webhook_enc_key))
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
	req.Header.Set("X-Webhook-Signature", utils.Sign(payload, secret))

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
