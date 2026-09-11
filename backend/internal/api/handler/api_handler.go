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
	Name  string `json:"name"`
	Email string `json:"email"`
}

type GetApiKeyBody struct{
	DeveloperId uuid.UUID `json:"developer_id"`
}

func (a *ApiHandler) AddApiKey(w http.ResponseWriter, r *http.Request) {
	data := RequestApiBody{}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		WriteError(w, http.StatusBadRequest, err)
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

	WriteJSON(w,http.StatusOK,data)
}

func (a *ApiHandler) GetApiKey(w http.ResponseWriter, r *http.Request) {
	var reqBody GetApiKeyBody;
	err:=json.NewDecoder(r.Body).Decode(&reqBody)
	if err!=nil{
		WriteError(w,http.StatusBadRequest,err)
		return
	}
	data, err := a.queries.GetUserApiKeys(r.Context(), reqBody.DeveloperId)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err)
		return
	}
	WriteJSON(w,http.StatusOK,data)
}
