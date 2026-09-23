package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Shubham-rawat0/temp-mail/SwiftInbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftInbox/backend/internal/service"
	"github.com/Shubham-rawat0/temp-mail/SwiftInbox/backend/internal/utils"
	"github.com/google/uuid"
)

type MailboxHandler struct {
	svc *service.MailboxService
}

type RequestEmailBody struct {
	Username  string     `json:"username"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

type mailboxResponse struct {
	Address   string        `json:"address"`
	CreatedAt time.Time     `json:"createdAt"`
	ExpiresAt time.Time     `json:"expiresAt"`
	CreatedBy uuid.NullUUID `json:"createdBy"`
}

func NewMailboxHandler(q *postgres.Queries) *MailboxHandler {
	return &MailboxHandler{
		svc: service.NewMailboxService(q),
	}
}

func NewMailboxHandlerWithService(svc *service.MailboxService) *MailboxHandler {
	return &MailboxHandler{
		svc: svc,
	}
}

// Public endpoints

func (m *MailboxHandler) CreateEmail(w http.ResponseWriter, r *http.Request) {
	reqBody := RequestEmailBody{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	result, err := m.svc.CreateCustomMailbox(r.Context(), reqBody.Username, nil, nil)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, mailboxResponse{
		Address:   result.Address,
		CreatedAt: result.CreatedAt,
		ExpiresAt: result.ExpiresAt,
		CreatedBy: result.CreatedBy,
	})
}

func (m *MailboxHandler) CreateMailbox(w http.ResponseWriter, r *http.Request) {
	reqBody := RequestEmailBody{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	result, err := m.svc.CreateMailbox(r.Context(), reqBody.Username, nil, nil)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, mailboxResponse{
		Address:   result.Address,
		CreatedAt: result.CreatedAt,
		ExpiresAt: result.ExpiresAt,
		CreatedBy: result.CreatedBy,
	})
}

// Developer Dashboard endpoints

func (m *MailboxHandler) CreateDeveloperEmail(w http.ResponseWriter, r *http.Request) {
	devID, ok := utils.DeveloperIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, errors.New("developer sign-in required"))
		return
	}

	reqBody := RequestEmailBody{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	result, err := m.svc.CreateCustomMailbox(r.Context(), reqBody.Username, &devID, reqBody.ExpiresAt)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, mailboxResponse{
		Address:   result.Address,
		CreatedAt: result.CreatedAt,
		ExpiresAt: result.ExpiresAt,
		CreatedBy: result.CreatedBy,
	})
}

func (m *MailboxHandler) CreateDeveloperMailbox(w http.ResponseWriter, r *http.Request) {
	devID, ok := utils.DeveloperIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, errors.New("developer sign-in required"))
		return
	}

	reqBody := RequestEmailBody{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	result, err := m.svc.CreateMailbox(r.Context(), reqBody.Username, &devID, reqBody.ExpiresAt)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, mailboxResponse{
		Address:   result.Address,
		CreatedAt: result.CreatedAt,
		ExpiresAt: result.ExpiresAt,
		CreatedBy: result.CreatedBy,
	})
}

func (m *MailboxHandler) ListDeveloperMailboxes(w http.ResponseWriter, r *http.Request) {
	devID, ok := utils.DeveloperIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, errors.New("developer sign-in required"))
		return
	}

	mailboxes, err := m.svc.ListDeveloperMailboxes(r.Context(), devID)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, mailboxes)
}

func (m *MailboxHandler) DeleteDeveloperMailbox(w http.ResponseWriter, r *http.Request) {
	devID, ok := utils.DeveloperIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, errors.New("developer sign-in required"))
		return
	}

	identifier := r.PathValue("identifier")
	if identifier == "" {
		WriteError(w, http.StatusBadRequest, errors.New("address is required"))
		return
	}

	result, err := m.svc.DeleteDeveloperMailbox(r.Context(), devID, identifier)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "mailbox deleted",
		"address": result.Address,
		"id":      result.ID.String(),
	})
}

// External API endpoints (/api/v1/*)

func (m *MailboxHandler) CreateAPIMailbox(w http.ResponseWriter, r *http.Request) {
	devID, ok := utils.DeveloperIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, errors.New("api key required"))
		return
	}

	reqBody := RequestEmailBody{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	result, err := m.svc.CreateMailbox(r.Context(), reqBody.Username, &devID, reqBody.ExpiresAt)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, mailboxResponse{
		Address:   result.Address,
		CreatedAt: result.CreatedAt,
		ExpiresAt: result.ExpiresAt,
		CreatedBy: result.CreatedBy,
	})
}

func (m *MailboxHandler) CreateCustomAPIMailbox(w http.ResponseWriter, r *http.Request) {
	devID, ok := utils.DeveloperIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, errors.New("api key required"))
		return
	}

	reqBody := RequestEmailBody{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	result, err := m.svc.CreateCustomMailbox(r.Context(), reqBody.Username, &devID, reqBody.ExpiresAt)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, mailboxResponse{
		Address:   result.Address,
		CreatedAt: result.CreatedAt,
		ExpiresAt: result.ExpiresAt,
		CreatedBy: result.CreatedBy,
	})
}

func (m *MailboxHandler) ListAPIMailboxes(w http.ResponseWriter, r *http.Request) {
	devID, ok := utils.DeveloperIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, errors.New("api key required"))
		return
	}

	mailboxes, err := m.svc.ListDeveloperMailboxes(r.Context(), devID)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, mailboxes)
}

func (m *MailboxHandler) GetAPIMailbox(w http.ResponseWriter, r *http.Request) {
	devID, ok := utils.DeveloperIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, errors.New("api key required"))
		return
	}

	idStr := r.PathValue("id")
	mailboxID, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, errors.New("invalid mailbox id"))
		return
	}

	mailbox, err := m.svc.GetDeveloperMailbox(r.Context(), devID, mailboxID)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, mailbox)
}

func (m *MailboxHandler) DeleteAPIMailbox(w http.ResponseWriter, r *http.Request) {
	devID, ok := utils.DeveloperIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, errors.New("api key required"))
		return
	}

	identifier := r.PathValue("id")
	if identifier == "" {
		identifier = r.PathValue("address")
	}

	result, err := m.svc.DeleteDeveloperMailbox(r.Context(), devID, identifier)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "mailbox deleted",
		"address": result.Address,
		"id":      result.ID.String(),
	})
}
