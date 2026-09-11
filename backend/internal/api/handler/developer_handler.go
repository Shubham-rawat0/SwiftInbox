package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type DeveloperHandler struct {
	queries *postgres.Queries
}

func NewDeveloperHandler(q *postgres.Queries) *DeveloperHandler {
	return &DeveloperHandler{
		queries: q,
	}
}

func (a *DeveloperHandler) CreateDeveloper(w http.ResponseWriter, r *http.Request) {
	var body CreateDeveloperBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	body.Name = strings.TrimSpace(body.Name)
	body.Email = strings.TrimSpace(body.Email)
	if body.Name == "" {
		WriteError(w, http.StatusBadRequest, errors.New("name is required"))
		return
	}

	address, err := mail.ParseAddress(body.Email)
	if err != nil || address.Address != body.Email {
		WriteError(w, http.StatusBadRequest, errors.New("valid email is required"))
		return
	}

	developer, err := a.queries.CreateDeveloper(
		r.Context(),
		postgres.CreateDeveloperParams{
			ID:    uuid.New(),
			Name:  body.Name,
			Email: body.Email,
		},
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			WriteError(w, http.StatusConflict, errors.New("email is already registered"))
			return
		}

		WriteError(w, http.StatusInternalServerError, errors.New("failed to create developer"))
		return
	}

	WriteJSON(w, http.StatusCreated, developer)
}

func (a *DeveloperHandler) GetDeveloper(w http.ResponseWriter, r *http.Request){
	
}