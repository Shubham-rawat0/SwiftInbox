package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Shubham-rawat0/temp-mail/SwiftInbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftInbox/backend/internal/service"
	"github.com/Shubham-rawat0/temp-mail/SwiftInbox/backend/internal/utils"
	"github.com/google/uuid"
)

type WebhookHandler struct {
	svc *service.WebhookService
}

func NewWebhookHandler(q *postgres.Queries) *WebhookHandler {
	return &WebhookHandler{
		svc: service.NewWebhookService(q),
	}
}

func NewWebhookHandlerWithService(svc *service.WebhookService) *WebhookHandler {
	return &WebhookHandler{
		svc: svc,
	}
}

type WebhookReqBody struct {
	Name       string      `json:"name"`
	Url        string      `json:"url"`
	Events     []string    `json:"events"`
	MailboxIDs []uuid.UUID `json:"mailbox_ids"`
}

type WebhookUpdateBody struct {
	Name     *string `json:"name"`
	IsActive *bool   `json:"is_active"`
}

type WebhookEventsBody struct {
	Events []string `json:"events"`
}

type WebhookMailboxBody struct {
	MailboxID uuid.UUID `json:"mailbox_id"`
}

func (a *WebhookHandler) getDeveloperID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	developerID, ok := utils.DeveloperIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return uuid.Nil, false
	}
	return developerID, true
}

func (a *WebhookHandler) getWebhookID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	webhookID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, errors.New("invalid webhook id"))
		return uuid.Nil, false
	}
	return webhookID, true
}

func (a *WebhookHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	developerID, ok := a.getDeveloperID(w, r)
	if !ok {
		return
	}

	reqBody := WebhookReqBody{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		WriteError(w, http.StatusBadRequest, errors.New("provide complete req body"))
		return
	}

	result, err := a.svc.CreateWebhook(r.Context(), developerID, reqBody.Name, reqBody.Url, reqBody.Events, reqBody.MailboxIDs)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func (a *WebhookHandler) UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	developerID, ok := a.getDeveloperID(w, r)
	if !ok {
		return
	}
	webhookID, ok := a.getWebhookID(w, r)
	if !ok {
		return
	}

	reqBody := WebhookUpdateBody{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		WriteError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	name := ""
	if reqBody.Name != nil {
		name = *reqBody.Name
	}

	isActive := true
	if reqBody.IsActive != nil {
		isActive = *reqBody.IsActive
	}

	result, err := a.svc.UpdateWebhook(r.Context(), developerID, webhookID, name, isActive)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func (a *WebhookHandler) AddMailbox(w http.ResponseWriter, r *http.Request) {
	developerID, ok := a.getDeveloperID(w, r)
	if !ok {
		return
	}
	webhookID, ok := a.getWebhookID(w, r)
	if !ok {
		return
	}

	var request WebhookMailboxBody
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.MailboxID == uuid.Nil {
		WriteError(w, http.StatusBadRequest, errors.New("mailbox_id is required"))
		return
	}

	mailboxID, err := a.svc.AddMailbox(r.Context(), developerID, webhookID, request.MailboxID)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, map[string]string{
		"message":    "mailbox added to webhook",
		"mailbox_id": mailboxID.String(),
	})
}

func (a *WebhookHandler) RemoveMailbox(w http.ResponseWriter, r *http.Request) {
	developerID, ok := a.getDeveloperID(w, r)
	if !ok {
		return
	}
	webhookID, ok := a.getWebhookID(w, r)
	if !ok {
		return
	}

	mailboxID, err := uuid.Parse(r.PathValue("mailboxID"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, errors.New("invalid mailbox id"))
		return
	}

	removedID, err := a.svc.RemoveMailbox(r.Context(), developerID, webhookID, mailboxID)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message":    "mailbox removed from webhook",
		"mailbox_id": removedID.String(),
	})
}

func (a *WebhookHandler) GetWebhook(w http.ResponseWriter, r *http.Request) {
	developerID, ok := a.getDeveloperID(w, r)
	if !ok {
		return
	}
	webhookID, ok := a.getWebhookID(w, r)
	if !ok {
		return
	}

	result, err := a.svc.GetWebhook(r.Context(), developerID, webhookID)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func (a *WebhookHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	developerID, ok := a.getDeveloperID(w, r)
	if !ok {
		return
	}

	webhooks, err := a.svc.ListWebhooks(r.Context(), developerID)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, webhooks)
}

func (a *WebhookHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	developerID, ok := a.getDeveloperID(w, r)
	if !ok {
		return
	}
	webhookID, ok := a.getWebhookID(w, r)
	if !ok {
		return
	}

	if err := a.svc.DeleteWebhook(r.Context(), developerID, webhookID); err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "webhook deleted",
		"id":      webhookID.String(),
	})
}

func (a *WebhookHandler) AddEvents(w http.ResponseWriter, r *http.Request) {
	developerID, ok := a.getDeveloperID(w, r)
	if !ok {
		return
	}
	webhookID, ok := a.getWebhookID(w, r)
	if !ok {
		return
	}

	request := WebhookEventsBody{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteError(w, http.StatusBadRequest, errors.New("invalid event body"))
		return
	}

	result, err := a.svc.AddEvents(r.Context(), developerID, webhookID, request.Events)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func (a *WebhookHandler) RemoveEvents(w http.ResponseWriter, r *http.Request) {
	developerID, ok := a.getDeveloperID(w, r)
	if !ok {
		return
	}
	webhookID, ok := a.getWebhookID(w, r)
	if !ok {
		return
	}

	request := WebhookEventsBody{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteError(w, http.StatusBadRequest, errors.New("invalid event body"))
		return
	}

	result, err := a.svc.RemoveEvents(r.Context(), developerID, webhookID, request.Events)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func (a *WebhookHandler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	developerID, ok := a.getDeveloperID(w, r)
	if !ok {
		return
	}
	webhookID, ok := a.getWebhookID(w, r)
	if !ok {
		return
	}

	statusCode, success, err := a.svc.TestWebhook(r.Context(), developerID, webhookID)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"success":    success,
		"statusCode": statusCode,
	})
}

func (a *WebhookHandler) GetDeadLetters(w http.ResponseWriter, r *http.Request) {
	developerID, ok := a.getDeveloperID(w, r)
	if !ok {
		return
	}

	deadLetters, err := a.svc.ListDeadLetters(r.Context(), developerID)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, deadLetters)
}

func (a *WebhookHandler) MarkDeadLetterSeen(w http.ResponseWriter, r *http.Request) {
	developerID, ok := a.getDeveloperID(w, r)
	if !ok {
		return
	}

	deadLetterID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, errors.New("invalid dead letter id"))
		return
	}

	id, err := a.svc.MarkDeadLetterSeen(r.Context(), developerID, deadLetterID)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"message": "dead letter marked as seen",
		"id":      id,
	})
}
