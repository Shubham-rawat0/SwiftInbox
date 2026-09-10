package utils

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const developerIDKey contextKey = "developer_id"
const apiKeyIDKey contextKey = "api_key_id"

func WithDeveloperID(ctx context.Context, developerID uuid.UUID) context.Context {
	return context.WithValue(ctx, developerIDKey, developerID)
}

func DeveloperIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	developerID, ok := ctx.Value(developerIDKey).(uuid.UUID)
	return developerID, ok
}

func WithAPIKeyID(ctx context.Context, apiKeyID uuid.UUID) context.Context {
	return context.WithValue(ctx, apiKeyIDKey, apiKeyID)
}

func APIKeyIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	apiKeyID, ok := ctx.Value(apiKeyIDKey).(uuid.UUID)
	return apiKeyID, ok
}

var usageCosts = map[string]int{
	"mailbox.create":  1,
	"message.list":    3,
	"message.get":     3,
	"attachment.list": 3,
	"attachment.get":  2,
	"mailbox.delete":  1,
}

const usageOperationKey contextKey = "usageOperation"

func UsageOperation(ctx context.Context) (string, int, bool) {
	operation, ok := ctx.Value(usageOperationKey).(string)
	if !ok {
		return "", 0, false
	}

	cost, ok := usageCosts[operation]
	return operation, cost, ok
}

func UsageCategory(operation string) string {
	switch operation {
	case "mailbox.create", "mailbox.delete":
		return "mailbox"
	case "message.list", "message.get", "attachment.get":
		return "message"
	default:
		return "api"
	}
}

func WithUsage(operation string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := usageCosts[operation]; !ok {
			http.Error(w, "unknown usage operation", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(
			r.Context(),
			usageOperationKey,
			operation,
		)

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
