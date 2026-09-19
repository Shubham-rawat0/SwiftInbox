package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/utils"
	"github.com/google/uuid"
)

var (
	ErrApiKeyNotFound = errors.New("api key not found")
)

type ApiKeyCreatedResult struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	ApiKey string    `json:"api_key"`
}

type ApiKeyService struct {
	queries *postgres.Queries
}

func NewApiKeyService(q *postgres.Queries) *ApiKeyService {
	return &ApiKeyService{
		queries: q,
	}
}

func (s *ApiKeyService) CreateApiKey(ctx context.Context, devID uuid.UUID, name string) (*ApiKeyCreatedResult, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "default"
	}

	id := uuid.New()
	apiKey, err := utils.GenerateAPIKey()
	if err != nil {
		return nil, err
	}

	keyHash := utils.HashAPIKey(apiKey)

	api, err := s.queries.CreateApiKey(ctx, postgres.CreateApiKeyParams{
		ID:          id,
		DeveloperID: devID,
		Name:        name,
		KeyHash:     keyHash,
		LastUsedAt: sql.NullTime{
			Valid: false,
		},
	})
	if err != nil {
		return nil, err
	}

	return &ApiKeyCreatedResult{
		ID:     api.ID,
		Name:   api.Name,
		ApiKey: apiKey,
	}, nil
}

func (s *ApiKeyService) ListApiKeys(ctx context.Context, devID uuid.UUID) ([]postgres.GetUserApiKeysRow, error) {
	return s.queries.GetUserApiKeys(ctx, devID)
}

func (s *ApiKeyService) RevokeDeveloperApiKey(ctx context.Context, devID uuid.UUID, keyID uuid.UUID) error {
	_, err := s.queries.RevokeDeveloperApiKey(ctx, postgres.RevokeDeveloperApiKeyParams{
		ID:          keyID,
		DeveloperID: devID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrApiKeyNotFound
		}
		return err
	}
	return nil
}

func (s *ApiKeyService) RevokeApiKeyByID(ctx context.Context, keyID uuid.UUID) error {
	_, err := s.queries.RevokeApiKey(ctx, keyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrApiKeyNotFound
		}
		return err
	}
	return nil
}

