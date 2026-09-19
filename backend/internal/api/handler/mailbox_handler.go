package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
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
	Username  string     `json:"username"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

type mailboxResponse struct {
	Address   string        `json:"address"`
	CreatedAt time.Time     `json:"createdAt"`
	ExpiresAt time.Time     `json:"expiresAt"`
	CreatedBy uuid.NullUUID `json:"createdBy"`
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
		WriteError(w, http.StatusInternalServerError, errors.New("failed to generate email address"))
		return
	}

	id := uuid.New()
	createdBy, ok := utils.DeveloperIDFromContext(r.Context())
	expiresAt, status, expiryErr := mailboxExpiresAt(reqBody.ExpiresAt, ok)
	if expiryErr != nil {
		WriteError(w, status, expiryErr)
		return
	}
	createdByValue := uuid.NullUUID{Valid: false}
	if ok {
		createdByValue = uuid.NullUUID{UUID: createdBy, Valid: true}
	}

	data, err := m.queries.CreateCustomEmailAddress(r.Context(), postgres.CreateCustomEmailAddressParams{ID: id, Address: address, ExpiresAt: expiresAt, CreatedBy: createdByValue})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(w, http.StatusConflict, errors.New("email address already exists"))
			return
		}
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

	WriteJSON(w, http.StatusCreated, mailboxResponse{
		Address:   data.Address,
		CreatedAt: data.CreatedAt,
		ExpiresAt: data.ExpiresAt,
		CreatedBy: data.CreatedBy,
	})
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

	id := uuid.New()
	createdBy, ok := utils.DeveloperIDFromContext(r.Context())
	expiresAt, status, expiryErr := mailboxExpiresAt(reqBody.ExpiresAt, ok)
	if expiryErr != nil {
		WriteError(w, status, expiryErr)
		return
	}
	createdByValue := uuid.NullUUID{Valid: false}
	if ok {
		createdByValue = uuid.NullUUID{UUID: createdBy, Valid: true}
	}

	mb, err := m.queries.CreateEmailAddress(
		r.Context(),
		postgres.CreateEmailAddressParams{
			ID:        id,
			Address:   address,
			ExpiresAt: expiresAt,
			CreatedBy: createdByValue,
		},
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(w, http.StatusNotFound, err)
			return
		}
		WriteError(w, http.StatusInternalServerError, errors.New("failed to create mailbox"))
		return
	}

	WriteJSON(w, http.StatusCreated, mailboxResponse{
		Address:   mb.Address,
		CreatedAt: mb.CreatedAt,
		ExpiresAt: mb.ExpiresAt,
		CreatedBy: mb.CreatedBy,
	})
}

func (m *MailboxHandler) DeleteMailbox(w http.ResponseWriter, r *http.Request) {
	address := r.PathValue("address")
	if address == "" {
		WriteError(w, http.StatusBadRequest, errors.New("address is required"))
	}
	address = utils.NormalizeAddress(address)
	if !utils.IsOurDomain(address) {
		WriteError(
			w,
			http.StatusBadRequest,
			errors.New("wrong domain"),
		)
		return
	}
	developerID, ok := utils.DeveloperIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, errors.New("api key required"))
		return
	}

	mb, err := m.queries.DeleteMailbox(r.Context(), postgres.DeleteMailboxParams{
		Address:   address,
		CreatedBy: uuid.NullUUID{UUID: developerID, Valid: true},
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(w, http.StatusNotFound, err)
			return
		}
		WriteError(w, http.StatusInternalServerError, errors.New("failed to delete mailbox"))
		return
	}

	WriteJSON(w, http.StatusCreated, map[string]string{
		"message": "mailbox deleted",
		"address": mb.Address,
		"id":      mb.ID.String(),
	})
}

func (m *MailboxHandler) ListDeveloperMailboxes(w http.ResponseWriter, r *http.Request){
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, errors.New("need developer id"))
		return
	}
	Id, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, errors.New("error parsing id"))
		return
	}
	data, err:=m.queries.ListDeveloperMailboxes(r.Context(),uuid.NullUUID{
		UUID:  Id,
		Valid: true,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(w, http.StatusNotFound, err)
			return
		}
		WriteError(w, http.StatusInternalServerError, errors.New("failed to get mailboxes"))
		return
	}

	WriteJSON(w,http.StatusOK,data)
}

func mailboxExpiresAt(requested *time.Time, isDeveloper bool) (time.Time, int, error) {
	if requested == nil {
		if isDeveloper {
			return time.Now().Add(30 * 24 * time.Hour), 0, nil
		}
		return time.Now().Add(24 * time.Hour), 0, nil
	}

	if !isDeveloper {
		return time.Time{}, http.StatusForbidden, errors.New("only developers can set expiresAt")
	}

	if !requested.After(time.Now()) {
		return time.Time{}, http.StatusBadRequest, errors.New("expiresAt must be in the future")
	}

	return requested.UTC(), 0, nil
}

