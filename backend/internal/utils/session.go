package utils

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const DeveloperCookieName = "developer_session"

var DeveloperCookieExpiry = time.Date(9999, time.December, 31, 23, 59, 59, 0, time.UTC)

func SignDeveloperCookie(developerID uuid.UUID, passwordHash string) string {
	return developerID.String() + "." + Sign([]byte(developerID.String()), passwordHash)
}

func ParseDeveloperCookie(value string) (uuid.UUID, string, error) {
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

