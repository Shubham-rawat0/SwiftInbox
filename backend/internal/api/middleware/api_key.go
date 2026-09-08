package middleware

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/api/handler"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/utils"
)

type ApiKeyMiddleware struct {
	queries *postgres.Queries
}

func NewApiMiddlewareHandler(q *postgres.Queries) *ApiKeyMiddleware {
	return &ApiKeyMiddleware{
		queries: q,
	}
}

const (
	apiKeyIDKey contextKey = "api_key_id"
)

type contextKey string

func (a *ApiKeyMiddleware) APIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		fmt.Println("apikey")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") ||
			parts[1] == "" {

			handler.WriteError(
				w,
				http.StatusUnauthorized,
				errors.New("invalid authorization header"),
			)
			return
		}

		apiKey := parts[1]
		keyHash := utils.HashAPIKey(apiKey)

		data, err := a.queries.GetApiKey(
			r.Context(),
			keyHash,
		)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				handler.WriteError(
					w,
					http.StatusUnauthorized,
					errors.New("invalid api key"),
				)
				return
			}

			handler.WriteError(
				w,
				http.StatusInternalServerError,
				errors.New("failed to authenticate api key"),
			)
			return
		}

		if data.RevokedAt.Valid {
			handler.WriteError(
				w,
				http.StatusUnauthorized,
				errors.New("api key has been revoked"),
			)
			return
		}

		ctx := utils.WithDeveloperID(r.Context(), data.DeveloperID)

		ctx = context.WithValue(
			ctx,
			apiKeyIDKey,
			data.ID,
		)

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (a *ApiKeyMiddleware) RequireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			handler.WriteError(
				w,
				http.StatusUnauthorized,
				errors.New("api key required"),
			)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") ||
			parts[1] == "" {

			handler.WriteError(
				w,
				http.StatusUnauthorized,
				errors.New("invalid authorization header"),
			)
			return
		}

		apiKey := parts[1]
		keyHash := utils.HashAPIKey(apiKey)

		data, err := a.queries.GetApiKey(
			r.Context(),
			keyHash,
		)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				handler.WriteError(
					w,
					http.StatusUnauthorized,
					errors.New("invalid api key"),
				)
				return
			}

			handler.WriteError(
				w,
				http.StatusInternalServerError,
				errors.New("failed to authenticate api key"),
			)
			return
		}

		if data.RevokedAt.Valid {
			handler.WriteError(
				w,
				http.StatusUnauthorized,
				errors.New("api key has been revoked"),
			)
			return
		}

		ctx := utils.WithDeveloperID(r.Context(), data.DeveloperID)

		ctx = context.WithValue(
			ctx,
			apiKeyIDKey,
			data.ID,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
