package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/utils"
	"github.com/google/uuid"
)

type MailboxHandler struct {
	queries *postgres.Queries
}

type RequestEmailBody struct {
	Username string `json:"username"`
}

func NewMailboxHandler(q *postgres.Queries) *MailboxHandler {
	return &MailboxHandler{
		queries: q,
	}
}

func (m *MailboxHandler) CreateEmail(w http.ResponseWriter, r *http.Request) {
	reqBody := RequestEmailBody{}

	err := json.NewDecoder(r.Body).Decode(&reqBody)

	if err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	if reqBody.Username == "" {
		WriteError(w, http.StatusBadRequest, errors.New("username is required"))
		return
	}

	address, err := utils.MakeCustomAddress(reqBody.Username)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	if address == "" {
		fmt.Println("makeAddress return empty for", reqBody.Username)
		WriteError(w, http.StatusBadGateway, err)
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	id := uuid.New()

	data, err := m.queries.CreateEmailAddress(r.Context(), postgres.CreateEmailAddressParams{ID: id, Address: address, ExpiresAt: expiresAt})

	if err != nil {
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

	WriteJSON(w, http.StatusCreated, data)
}

func (m *MailboxHandler) CreateMailbox(w http.ResponseWriter, r *http.Request) {
	reqBody := &RequestEmailBody{}

	if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	if reqBody.Username == "" {
		WriteError(
			w,
			http.StatusBadRequest,
			errors.New("address required"),
		)
		return
	}

	address := utils.NormalizeAddress(reqBody.Username)

	if !utils.IsOurDomain(address) {
		WriteError(
			w,
			http.StatusBadRequest,
			errors.New("wrong domain"),
		)
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	id := uuid.New()

	mb, err := m.queries.CreateEmailAddress(
		r.Context(),
		postgres.CreateEmailAddressParams{
			ID:        id,
			Address:   address,
			ExpiresAt: expiresAt,
		},
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(w, http.StatusNotFound, err)
			return
		}
		WriteError(w,http.StatusInternalServerError,errors.New("failed to create mailbox"),)
		return
	}

	WriteJSON(w, http.StatusCreated, map[string]string{
		"address": mb.Address,
	})
}


