package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/parser"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/utils"
	"github.com/google/uuid"
)

type MessageHandler struct {
    queries *postgres.Queries
}

type ParseEmail struct{
	ID   uuid.UUID    `json:"id"`
	Sender   string  `json:"sender"` 
	Subject   string  `json:"subject"` 
	Raw      *parser.ParsedEmail  `json:"raw"`
	CreatedAt  time.Time    `json:"created_at"`
}

type RequestRawMessage struct{
	ID   uuid.UUID    `json:"id"`
	Address  string  `json:"address"`
}

func NewMessageHandler(q *postgres.Queries) *MessageHandler{
	return &MessageHandler{
		queries:q,
	}
}

func (m *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request){
	address:=r.PathValue("address")
	address=utils.NormalizeAddress(address)

	data,err:=m.queries.GetMessages(r.Context(),address)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(w, http.StatusNotFound, errors.New("message not found"))
			return
		}
		WriteError(w,http.StatusInternalServerError,errors.New("failed to create mailbox"),)
		return
	}

	WriteJSON(w, http.StatusCreated, data)
}

func (m *MessageHandler) GetRawMessage(w http.ResponseWriter, r *http.Request) {
	reqBody := RequestRawMessage{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		WriteError(w, http.StatusBadRequest, errors.New("provide proper body"))
		return
	}

	data, err := m.queries.GetRawMessage(r.Context(), postgres.GetRawMessageParams{
		ID:      reqBody.ID,
		Address: reqBody.Address,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(w, http.StatusNotFound, errors.New("message not found"))
			return
		}
		WriteError(w, http.StatusInternalServerError, errors.New("failed to fetch raw message"))
		return
	}

	parsedEmail, err := parser.ParseRawMail(data.Raw)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, errors.New("failed to parse raw message"))
		return
	}

	resp := ParseEmail{
		ID:        data.ID,
		Subject:   data.Subject.String,
		Sender:    data.Sender, 
		CreatedAt: data.CreatedAt,
		Raw:       parsedEmail,
	}

	WriteJSON(w, http.StatusOK, resp)
}