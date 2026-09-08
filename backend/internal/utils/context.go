package utils

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const developerIDKey contextKey = "developer_id"

func WithDeveloperID(ctx context.Context, developerID uuid.UUID) context.Context {
	return context.WithValue(ctx, developerIDKey, developerID)
}

func DeveloperIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	developerID, ok := ctx.Value(developerIDKey).(uuid.UUID)
	return developerID, ok
}
