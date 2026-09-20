package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/service"
)

func WriteJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, code int, err error) {
	WriteJSON(w, code, map[string]string{
		"error": err.Error(),
	})
}

func WriteServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrMailboxNotFound),
		errors.Is(err, service.ErrMessageNotFound),
		errors.Is(err, service.ErrAttachmentNotFound),
		errors.Is(err, service.ErrWebhookNotFound),
		errors.Is(err, service.ErrWebhookMailboxNotFound),
		errors.Is(err, service.ErrWebhookMailboxLinkNotFound),
		errors.Is(err, service.ErrApiKeyNotFound):
		WriteError(w, http.StatusNotFound, err)
	case errors.Is(err, service.ErrMailboxAlreadyExists):
		WriteError(w, http.StatusConflict, err)
	case errors.Is(err, service.ErrForbiddenDeveloperMailbox):
		WriteError(w, http.StatusForbidden, err)
	case errors.Is(err, service.ErrWebhookInactive):
		WriteError(w, http.StatusConflict, err)
	case errors.Is(err, service.ErrWrongDomain),
		errors.Is(err, service.ErrInvalidAddress),
		errors.Is(err, service.ErrInvalidExpiry),
		errors.Is(err, service.ErrUsernameRequired),
		errors.Is(err, service.ErrAddressRequired),
		errors.Is(err, service.ErrEventsRequired),
		errors.Is(err, service.ErrMailboxIDsRequired),
		errors.Is(err, service.ErrInvalidAttachmentIndex),
		errors.Is(err, service.ErrMailboxNotOwned):
		WriteError(w, http.StatusBadRequest, err)
	case errors.Is(err, service.ErrInvalidWebhookEvent):
		WriteError(w, http.StatusBadRequest, err)
	case errors.Is(err, service.ErrMailboxAlreadyLinked):
		WriteError(w, http.StatusConflict, err)

	case errors.Is(err, service.ErrUnauthorized):
		WriteError(w, http.StatusUnauthorized, err)
	default:
		WriteError(w, http.StatusInternalServerError, err)
	}
}