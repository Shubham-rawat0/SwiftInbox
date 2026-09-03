package handler

import (
	"net/http"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
)

type MessageHandler struct {
    queries *postgres.Queries
}

func NewMessageHandler(q *postgres.Queries) *MessageHandler{
	return &MessageHandler{
		queries:q,
	}
}

func (m *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request){
	
}