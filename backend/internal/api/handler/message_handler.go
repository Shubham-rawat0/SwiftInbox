package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/service"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/utils"
	"github.com/google/uuid"
)

type MessageHandler struct {
	svc *service.MailboxService
}

func NewMessageHandler(q *postgres.Queries) *MessageHandler {
	return &MessageHandler{
		svc: service.NewMailboxService(q),
	}
}

func NewMessageHandlerWithService(svc *service.MailboxService) *MessageHandler {
	return &MessageHandler{
		svc: svc,
	}
}

func (m *MessageHandler) developerIDFromContext(r *http.Request) *uuid.UUID {
	devID, ok := utils.DeveloperIDFromContext(r.Context())
	if !ok {
		return nil
	}
	return &devID
}

func (m *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	address := r.PathValue("address")
	if address == "" {
		WriteError(w, http.StatusBadRequest, errors.New("address is required"))
		return
	}

	result, err := m.svc.GetMessages(r.Context(), address, m.developerIDFromContext(r))
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func (m *MessageHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		WriteError(w, http.StatusBadRequest, errors.New("need a message id"))
		return
	}

	messageID, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, errors.New("invalid message id"))
		return
	}

	result, err := m.svc.GetMessage(r.Context(), messageID, m.developerIDFromContext(r))
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func (m *MessageHandler) GetAttachment(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	indexStr := r.PathValue("index")

	if idStr == "" {
		WriteError(w, http.StatusBadRequest, errors.New("need a message id"))
		return
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, errors.New("invalid attachment index"))
		return
	}
	if index < 0 {
		WriteError(w, http.StatusBadRequest, errors.New("index should be greater than or equal to 0"))
		return
	}

	messageID, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, errors.New("invalid message id"))
		return
	}

	att, err := m.svc.GetAttachment(r.Context(), messageID, index, m.developerIDFromContext(r))
	if err != nil {
		WriteServiceError(w, err)
		return
	}

	if att.ContentType != "" {
		w.Header().Set("Content-Type", att.ContentType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	if att.Inline {
		// Inline parts render directly in the browser (e.g. embedded images).
		// They can still be saved via the frontend's download control.
		if att.Filename != "" {
			w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, att.Filename))
		} else {
			w.Header().Set("Content-Disposition", "inline")
		}
	} else if att.Filename != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, att.Filename))
	}

	if att.ContentID != "" {
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(att.Data); err != nil {
		log.Printf("failed to write attachment: %v", err)
	}
}
