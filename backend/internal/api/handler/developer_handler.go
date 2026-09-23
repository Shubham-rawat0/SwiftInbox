package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/Shubham-rawat0/temp-mail/SwiftInbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftInbox/backend/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type DeveloperHandler struct {
	queries *postgres.Queries
}

func NewDeveloperHandler(q *postgres.Queries) *DeveloperHandler {
	return &DeveloperHandler{
		queries: q,
	}
}

type CreateDeveloperBody struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignInDeveloperBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *DeveloperHandler) CreateDeveloper(w http.ResponseWriter, r *http.Request) {
	var body CreateDeveloperBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	body.Name = strings.TrimSpace(body.Name)
	body.Email = strings.TrimSpace(body.Email)
	if body.Name == "" {
		WriteError(w, http.StatusBadRequest, errors.New("name is required"))
		return
	}

	address, err := mail.ParseAddress(body.Email)
	if err != nil || address.Address != body.Email {
		WriteError(w, http.StatusBadRequest, errors.New("valid email is required"))
		return
	}
	if len(body.Password) < 8 {
		WriteError(w, http.StatusBadRequest, errors.New("password must be at least 8 characters"))
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, errors.New("failed to secure password"))
		return
	}

	developer, err := a.queries.CreateDeveloper(
		r.Context(),
		postgres.CreateDeveloperParams{
			ID:           uuid.New(),
			Name:         body.Name,
			Email:        body.Email,
			PasswordHash: string(passwordHash),
		},
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			WriteError(w, http.StatusConflict, errors.New("email is already registered"))
			return
		}

		WriteError(w, http.StatusInternalServerError, errors.New("failed to create developer"))
		return
	}

	WriteJSON(w, http.StatusCreated, developer)
}

func (a *DeveloperHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	logger := slog.Default()

	var body SignInDeveloperBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logger.WarnContext(r.Context(),
			"developer sign-in: invalid request body",
			"error", err,
		)

		WriteError(w, http.StatusBadRequest, err)
		return
	}

	email := strings.TrimSpace(body.Email)
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		logger.WarnContext(r.Context(),
			"developer sign-in: invalid email",
		)

		WriteError(w, http.StatusBadRequest, errors.New("valid email is required"))
		return
	}

	if body.Password == "" {
		logger.WarnContext(r.Context(),
			"developer sign-in: password missing",
		)

		WriteError(w, http.StatusBadRequest, errors.New("password is required"))
		return
	}

	developer, err := a.queries.GetDeveloperAuthByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.WarnContext(r.Context(),
				"developer sign-in: invalid credentials",
			)

			WriteError(w, http.StatusUnauthorized,
				errors.New("invalid email or password"))
			return
		}

		logger.ErrorContext(r.Context(),
			"developer sign-in: database error",
			"error", err,
		)

		WriteError(w, http.StatusInternalServerError,
			errors.New("failed to sign in"))
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(developer.PasswordHash),
		[]byte(body.Password),
	); err != nil {
		logger.WarnContext(r.Context(),
			"developer sign-in: invalid credentials",
			"developer_id", developer.ID,
		)

		WriteError(w, http.StatusUnauthorized,
			errors.New("invalid email or password"))
		return
	}

	cookieValue := utils.SignDeveloperCookie(
		developer.ID,
		developer.PasswordHash,
	)

	http.SetCookie(w, &http.Cookie{
		Name:     utils.DeveloperCookieName,
		Value:    cookieValue,
		Path:     "/",
		Expires:  utils.DeveloperCookieExpiry,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	logger.InfoContext(r.Context(),
		"developer signed in",
		"developer_id", developer.ID,
	)

	WriteJSON(w, http.StatusOK, map[string]any{
		"id":    developer.ID,
		"name":  developer.Name,
		"email": developer.Email,
	})
}

func (a *DeveloperHandler) SignOut(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     utils.DeveloperCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	WriteJSON(w, http.StatusOK, map[string]string{"message": "signed out"})
}

func (a *DeveloperHandler) GetDeveloper(w http.ResponseWriter, r *http.Request) {
	devID, ok := utils.DeveloperIDFromContext(r.Context())
	if !ok {
		cookie, err := r.Cookie(utils.DeveloperCookieName)
		if err != nil {
			WriteError(w, http.StatusUnauthorized, errors.New("developer sign-in required"))
			return
		}

		cookieID, signature, err := utils.ParseDeveloperCookie(cookie.Value)
		if err != nil {
			WriteError(w, http.StatusUnauthorized, errors.New("invalid developer session"))
			return
		}

		passwordHash, err := a.queries.GetDeveloperPasswordHash(r.Context(), cookieID)
		if err != nil || !utils.Verify([]byte(cookieID.String()), passwordHash, signature) {
			WriteError(w, http.StatusUnauthorized, errors.New("invalid developer session"))
			return
		}
		devID = cookieID
	}

	pathID := r.PathValue("id")
	if pathID != "" && pathID != "me" {
		targetID, err := uuid.Parse(pathID)
		if err != nil {
			WriteError(w, http.StatusBadRequest, errors.New("invalid developer id"))
			return
		}
		if targetID != devID {
			WriteError(w, http.StatusUnauthorized, errors.New("invalid developer session"))
			return
		}
	}

	data, err := a.queries.GetDeveloper(r.Context(), devID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(w, http.StatusNotFound, errors.New("developer not found"))
			return
		}

		WriteError(w, http.StatusInternalServerError, errors.New("failed to get developer"))
		return
	}

	WriteJSON(w, http.StatusOK, data)
}
