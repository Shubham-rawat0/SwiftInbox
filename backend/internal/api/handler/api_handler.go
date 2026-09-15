package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/utils"
)

type ApiHandler struct {
	queries *postgres.Queries
}

func NewApiHandler(q *postgres.Queries) *ApiHandler {
	return &ApiHandler{
		queries: q,
	}
}

type RequestApiBody struct {
	DeveloperId uuid.UUID `json:"developer_id"`
	Name        string    `json:"name"`
}

type CreateDeveloperBody struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type GetApiKeyBody struct {
	DeveloperId uuid.UUID `json:"developer_id"`
}

func (a *ApiHandler) AddApiKey(w http.ResponseWriter, r *http.Request) {
	data := RequestApiBody{}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	cookie, err := r.Cookie(developerCookieName)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, errors.New("developer sign-in required"))
		return
	}

	cookieID, signature, err := parseDeveloperCookie(cookie.Value)
	if err != nil || cookieID != data.DeveloperId {
		WriteError(w, http.StatusUnauthorized, errors.New("invalid developer session"))
		return
	}

	passwordHash, err := a.queries.GetDeveloperPasswordHash(r.Context(), cookieID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(w, http.StatusUnauthorized, errors.New("invalid developer session"))
			return
		}

		WriteError(w, http.StatusInternalServerError, errors.New("failed to authenticate developer"))
		return
	}

	if !utils.Verify([]byte(cookieID.String()), passwordHash, signature) {
		WriteError(w, http.StatusUnauthorized, errors.New("invalid developer session"))
		return
	}

	data.Name = strings.TrimSpace(data.Name)
	if data.Name == "" {
		data.Name = "default"
	}

	id := uuid.New()

	apiKey, err := utils.GenerateAPIKey()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

	keyHash := utils.HashAPIKey(apiKey)

	api, err := a.queries.CreateApiKey(
		r.Context(),
		postgres.CreateApiKeyParams{
			ID:          id,
			DeveloperID: data.DeveloperId,
			Name:        data.Name,
			KeyHash:     keyHash,
			LastUsedAt: sql.NullTime{
				Valid: false,
			},
		},
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(w, http.StatusNotFound, err)
			return
		}

		WriteError(
			w,
			http.StatusInternalServerError,
			errors.New("failed to create api key"),
		)
		return
	}

	WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":      api.ID,
		"name":    api.Name,
		"api_key": apiKey,
	})
}

func (a *ApiHandler) RevokeApiKey(w http.ResponseWriter, r *http.Request) {
	id, ok := utils.APIKeyIDFromContext(r.Context())
	if !ok {
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

func (a *ApiHandler) GetApiKeyUsage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	devID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, http.StatusBadRequest, errors.New("invalid developer ID"))
		return
	}

	cookie, err := r.Cookie(developerCookieName)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, errors.New("developer sign-in required"))
		return
	}

	cookieID, signature, err := parseDeveloperCookie(cookie.Value)
	if err != nil || cookieID != devID {
		WriteError(w, http.StatusUnauthorized, errors.New("invalid developer session"))
		return
	}

	passwordHash, err := a.queries.GetDeveloperPasswordHash(r.Context(), devID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(w, http.StatusUnauthorized, errors.New("invalid developer session"))
			return
		}

		WriteError(w, http.StatusInternalServerError, errors.New("failed to authenticate developer"))
		return
	}

	if !utils.Verify([]byte(cookieID.String()), passwordHash, signature) {
		WriteError(w, http.StatusUnauthorized, errors.New("invalid developer session"))
		return
	}

	data, err := a.queries.GetUserApiKeys(r.Context(), devID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err)
		return
	}
	WriteJSON(w, http.StatusOK, data)
}
