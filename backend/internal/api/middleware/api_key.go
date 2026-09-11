package middleware

import (
	"context"
	"database/sql"
	"errors"
	"log"
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

func (a *ApiKeyMiddleware) APIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := getAuthHeader(r)
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		a.RequireAPIKey(next).ServeHTTP(w, r)
	})
}

func (a *ApiKeyMiddleware) RequireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := getAuthHeader(r)

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

		ctx = utils.WithAPIKeyID(ctx, data.ID)

		if err := a.recordUsage(ctx); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				handler.WriteError(w, http.StatusTooManyRequests, errors.New("usage quota exceeded"))
				return
			}
			log.Printf("failed to record api usage: %v", err)
			handler.WriteError(w, http.StatusInternalServerError, errors.New("failed to record api usage"))
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *ApiKeyMiddleware) recordUsage(ctx context.Context) error {
	developerID, ok := utils.DeveloperIDFromContext(ctx)
	if !ok {
		return nil
	}

	apiKeyID, ok := utils.APIKeyIDFromContext(ctx)
	if !ok {
		return nil
	}

	operation, cost, ok := utils.UsageOperation(ctx)
	if !ok {
		return nil
	}

	_, err := a.queries.IncrementUsage(ctx, postgres.IncrementUsageParams{
		DeveloperID: developerID,
		ApiRequests: int32(cost),
		Category:    utils.UsageCategory(operation),
	})
	if err != nil {
		return err
	}

	_, err = a.queries.IncrementApiKeyUsage(ctx, postgres.IncrementApiKeyUsageParams{
		ApiKeyID:    apiKeyID,
		ApiRequests: int32(cost),
		Category:    utils.UsageCategory(operation),
	})
	return err
}

func getAuthHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		authHeader = r.Header.Get("Authentication")
	}
	return authHeader
}
