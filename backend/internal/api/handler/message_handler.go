package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/utils"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/parser"
	"github.com/google/uuid"
)

type MessageHandler struct {
    queries *postgres.Queries
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
		WriteError(w,http.StatusInternalServerError,errors.New("failed to create mailbox"),)
		return
	}

	WriteJSON(w, http.StatusCreated, data)
}

func (m *MessageHandler) GetRawMessage(w http.ResponseWriter, r *http.Request){
	reqBody:=RequestRawMessage{}
	err:=json.NewDecoder(r.Body).Decode(&reqBody)
	if err!=nil{
		WriteError(w,http.StatusBadRequest,errors.New("Provide proper body"),)
		return
	}
	data, err:=m.queries.GetRawMessage(r.Context(),postgres.GetRawMessageParams{ID: reqBody.ID,Address: reqBody.Address})
	if err != nil {
		WriteError(w,http.StatusInternalServerError,errors.New("failed to fetch raw message"),)
		return
	}

	parsedEmail,err:=parser.ParseRawMail(data.Raw)
	if err!=nil{
		WriteError(w,http.StatusInternalServerError,err)
		return
	}

	WriteJSON(w,http.StatusOK,parsedEmail)
}