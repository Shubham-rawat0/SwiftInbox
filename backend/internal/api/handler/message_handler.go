package handler

import "github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"

type MessageHandler struct {
    queries *postgres.Queries
}

func NewMessageHandler(q *postgres.Queries) *MessageHandler{
	return &MessageHandler{
		queries:q,
	}
}