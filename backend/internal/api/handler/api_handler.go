package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/Shubham-rawat0/temp-mail/SwiftInbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftInbox/backend/internal/service"
	"github.com/Shubham-rawat0/temp-mail/SwiftInbox/backend/internal/utils"
)

type ApiHandler struct {
	apiKeyService  *service.ApiKeyService
	webhookService *service.WebhookService
	queries        *postgres.Queries
}

func NewApiHandler(q *postgres.Queries) *ApiHandler {
	return &ApiHandler{
		apiKeyService:  service.NewApiKeyService(q),
		webhookService: service.NewWebhookService(q),
		queries:        q,
	}
}

func NewApiHandlerWithServices(apiKeyService *service.ApiKeyService, webhookService *service.WebhookService, q *postgres.Queries) *ApiHandler {
	return &ApiHandler{
		apiKeyService:  apiKeyService,
		webhookService: webhookService,
		queries:        q,
	}
}

type RequestApiBody struct {
	DeveloperId uuid.UUID `json:"developer_id"`
	Name        string    `json:"name"`
}

func (a *ApiHandler) resolveDeveloperID(r *http.Request) (uuid.UUID, error) {
	if devID, ok := utils.DeveloperIDFromContext(r.Context()); ok {
		return devID, nil
	}

	cookie, err := r.Cookie(utils.DeveloperCookieName)
	if err != nil {
		return uuid.Nil, errors.New("developer sign-in required")
	}

	cookieID, signature, err := utils.ParseDeveloperCookie(cookie.Value)
	if err != nil {
		return uuid.Nil, errors.New("invalid developer session")
	}

	passwordHash, err := a.queries.GetDeveloperPasswordHash(r.Context(), cookieID)
	if err != nil || !utils.Verify([]byte(cookieID.String()), passwordHash, signature) {
		return uuid.Nil, errors.New("invalid developer session")
	}

	return cookieID, nil
}

func (a *ApiHandler) AddApiKey(w http.ResponseWriter, r *http.Request) {
	devID, err := a.resolveDeveloperID(r)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, err)
		return
	}

	data := RequestApiBody{}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&data)
	}

	if data.DeveloperId != uuid.Nil && data.DeveloperId != devID {
		WriteError(w, http.StatusUnauthorized, errors.New("invalid developer session"))
		return
	}

	result, err := a.apiKeyService.CreateApiKey(r.Context(), devID, data.Name)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, map[string]any{
		"id":      result.ID,
		"name":    result.Name,
		"api_key": result.ApiKey,
	})
}

func (a *ApiHandler) GetApiKeyUsage(w http.ResponseWriter, r *http.Request) {
	devID, err := a.resolveDeveloperID(r)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, err)
		return
	}

	pathID := r.PathValue("id")
	if pathID != "" && pathID != "me" {
		targetID, err := uuid.Parse(pathID)
		if err == nil && targetID != devID {
			WriteError(w, http.StatusUnauthorized, errors.New("invalid developer session"))
			return
		}
	}

	data, err := a.apiKeyService.ListApiKeys(r.Context(), devID)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, data)
}

func (a *ApiHandler) RevokeDeveloperApiKey(w http.ResponseWriter, r *http.Request) {
	devID, err := a.resolveDeveloperID(r)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, err)
		return
	}

	idStr := r.PathValue("id")
	keyID, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, errors.New("invalid api key id"))
		return
	}

	if err := a.apiKeyService.RevokeDeveloperApiKey(r.Context(), devID, keyID); err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "api key revoked",
		"id":      keyID.String(),
	})
}

func (a *ApiHandler) RevokeApiKey(w http.ResponseWriter, r *http.Request) {
	id, ok := utils.APIKeyIDFromContext(r.Context())
	if !ok {
		// If called from developer dashboard with path /api/dev/keys/{id}
		if pathID := r.PathValue("id"); pathID != "" {
			a.RevokeDeveloperApiKey(w, r)
			return
		}
		WriteError(w, http.StatusInternalServerError, errors.New("invalid api key id"))
		return
	}

	data, err := a.queries.RevokeApiKey(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

	WriteJSON(w, http.StatusOK, data)
}

func (a *ApiHandler) GetWebhookDeadLetters(w http.ResponseWriter, r *http.Request) {
	devID, ok := utils.DeveloperIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	data, err := a.webhookService.ListDeadLetters(r.Context(), devID)
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, data)
}
