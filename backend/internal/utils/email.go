package utils

import (
	"fmt"
	"os"
	"strings"
)

var allowedDomain = getAllowedDomain()

func getAllowedDomain() string {
	if domain := os.Getenv("SMTP_DOMAIN"); domain != "" {
		return domain
	}

	if domain := os.Getenv("MAIL_DOMAIN"); domain != "" {
		return domain
	}

	return "temp.mail.at"
}

func MakeCustomAddress(username string) (string, error) {
	username = strings.TrimSpace(username)
	username = strings.ToLower(username)

	var sanitized strings.Builder

	for _, r := range username {
		if (r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') ||
			r == '.' ||
			r == '-' ||
			r == '_' {
			sanitized.WriteRune(r)
		}
	}

	sanitizedUsername := sanitized.String()

	if sanitizedUsername == "" {
		return "", fmt.Errorf("invalid username")
	}

	if len(sanitizedUsername) < 1 || len(sanitizedUsername) > 64 {
		return "", fmt.Errorf("username must be between 1 and 64 characters")
	}

	return sanitizedUsername + "@" + allowedDomain, nil
}

func NormalizeAddress(addr string) string {
	return strings.ToLower(strings.TrimSpace(addr))
}

func ExtractDomain(addr string) string {
	if addr == "" {
		return ""
	}

	cleaned := strings.TrimSpace(addr)
	cleaned = strings.Trim(cleaned, "<>")

	atIndex := strings.LastIndex(cleaned, "@")
	if atIndex == -1 {
		return ""
	}

	domain := strings.ToLower(cleaned[atIndex+1:])

	domain = strings.TrimSuffix(domain, ".")

	return domain
}

func IsOurDomain(addr string) (bool) {
	domain := ExtractDomain(addr)
	allowed := strings.ToLower(allowedDomain)

	fmt.Printf("[DOMAIN CHECK] address=%s extractedDomain=%s allowedDomain=%s matches=%t\n",
		addr,
		domain,
		allowed,
		domain == allowed,
	)

	return domain == allowed 
}
