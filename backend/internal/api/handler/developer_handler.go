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

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type DeveloperHandler struct {
	queries *postgres.Queries
}

const developerCookieName = "developer_session"

var developerCookieExpiry = time.Date(9999, time.December, 31, 23, 59, 59, 0, time.UTC)

func NewDeveloperHandler(q *postgres.Queries) *DeveloperHandler {
	return &DeveloperHandler{
		queries: q,
	}
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

type SignInDeveloperBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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

	cookieValue := signDeveloperCookie(
		developer.ID,
		developer.PasswordHash,
	)

	http.SetCookie(w, &http.Cookie{
		Name:     developerCookieName,
		Value:    cookieValue,
		Path:     "/",
		Expires:  developerCookieExpiry,
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
		Name:     developerCookieName,
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
	id := r.PathValue("id")

	devID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, http.StatusBadRequest, errors.New("invalid developer ID"))
		return
	}

	cookie, err := r.Cookie(developerCookieName)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, errors.New("developer sign-in required"))
		return
	}

	cookieID, signature, err := parseDeveloperCookie(cookie.Value)
	if err != nil || cookieID != devID {
		WriteError(w, http.StatusUnauthorized, errors.New("invalid developer session"))
		return
	}

	passwordHash, err := a.queries.GetDeveloperPasswordHash(r.Context(), devID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(w, http.StatusUnauthorized, errors.New("invalid developer session"))
			return
		}
		WriteError(w, http.StatusInternalServerError, errors.New("failed to authenticate developer"))
		return
	}
	if !utils.Verify([]byte(cookieID.String()), passwordHash, signature) {
		WriteError(w, http.StatusUnauthorized, errors.New("invalid developer session"))
		return
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

func signDeveloperCookie(developerID uuid.UUID, passwordHash string) string {
	return developerID.String() + "." + utils.Sign([]byte(developerID.String()), passwordHash)
}

func parseDeveloperCookie(value string) (uuid.UUID, string, error) {
	developerIDValue, signature, ok := strings.Cut(value, ".")
	if !ok || signature == "" {
		return uuid.Nil, "", errors.New("invalid developer session")
	}

	developerID, err := uuid.Parse(developerIDValue)
	if err != nil {
		return uuid.Nil, "", errors.New("invalid developer session")
	}

	return developerID, signature, nil
}
