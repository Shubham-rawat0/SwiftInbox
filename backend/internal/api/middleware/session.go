package middleware

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/api/handler"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/utils"
)

type SessionMiddleware struct {
	queries *postgres.Queries
}

func NewSessionMiddleware(q *postgres.Queries) *SessionMiddleware {
	return &SessionMiddleware{
		queries: q,
	}
}

func (s *SessionMiddleware) RequireDeveloperSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(utils.DeveloperCookieName)
		if err != nil {
			handler.WriteError(w, http.StatusUnauthorized, errors.New("developer sign-in required"))
			return
		}

		devID, signature, err := utils.ParseDeveloperCookie(cookie.Value)
		if err != nil {
			handler.WriteError(w, http.StatusUnauthorized, errors.New("invalid developer session"))
			return
		}

		passwordHash, err := s.queries.GetDeveloperPasswordHash(r.Context(), devID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				handler.WriteError(w, http.StatusUnauthorized, errors.New("invalid developer session"))
				return
			}
			handler.WriteError(w, http.StatusInternalServerError, errors.New("failed to authenticate developer session"))
			return
		}

		if !utils.Verify([]byte(devID.String()), passwordHash, signature) {
			handler.WriteError(w, http.StatusUnauthorized, errors.New("invalid developer session"))
			return
		}

		ctx := utils.WithDeveloperID(r.Context(), devID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

