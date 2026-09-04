package handler

import (
	"database/sql"
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

type MessagePreview struct {
	ID        uuid.UUID `json:"id"`
	Sender    string    `json:"sender"`
	Subject   string    `json:"subject"`
	Preview   string    `json:"preview"`
	CreatedAt time.Time `json:"createdAt"`
}

type MessageResponse struct {
	Address      string           `json:"address"`
	Messages     []MessagePreview `json:"messages"`
	MessageCount int              `json:"messageCount"`
}

func NewMessageHandler(q *postgres.Queries) *MessageHandler {
	return &MessageHandler{
		queries: q,
	}
}

func (m *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	identifier := r.PathValue("address")
	address := utils.NormalizeAddress(identifier)

	var data []postgres.GetMessagesRow
	var err error
	
	data, err = m.queries.GetMessages(r.Context(), address)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(
				w,
				http.StatusNotFound,
				errors.New("mailbox not found"),
			)
			return
		}

		WriteError(
			w,
			http.StatusInternalServerError,
			errors.New("failed to fetch messages"),
		)
		return
	}

	messages := make([]MessagePreview, 0, len(data))

	for _, msg := range data {
		preview, err := parser.ParseTextBody(msg.Raw)

		if err != nil {
			preview = "Unable to load preview"
		}

		preview = parser.MakePreview(preview, 150)

		messages = append(messages, MessagePreview{
			ID:        msg.ID,
			Sender:    msg.Sender,
			Subject:   msg.Subject.String,
			Preview:   preview,
			CreatedAt: msg.CreatedAt,
		})
	}

	response := MessageResponse{
		Address:      address,
		Messages:     messages,
		MessageCount: len(messages),
	}

	WriteJSON(w, http.StatusOK, response)
}
