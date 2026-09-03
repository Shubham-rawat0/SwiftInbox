package handler

import (
	"net/http"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
)

type MailboxHandler struct{
	queries		*postgres.Queries
}

func NewMailboxHandler(q *postgres.Queries) *MailboxHandler{
	return &MailboxHandler{
		queries:q,
	}
}


func (m *MailboxHandler) CreateEmail(w http.ResponseWriter, r *http.Request){

}