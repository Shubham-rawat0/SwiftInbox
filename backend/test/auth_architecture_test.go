package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Shubham-rawat0/temp-mail/SwiftInbox/backend/internal/api/router"
	"github.com/Shubham-rawat0/temp-mail/SwiftInbox/backend/internal/utils"
	"github.com/google/uuid"
)

func TestSessionCookieHelpers(t *testing.T) {
	devID := uuid.New()
	passwordHash := "some-hashed-password"

	cookieVal := utils.SignDeveloperCookie(devID, passwordHash)
	parsedID, sig, err := utils.ParseDeveloperCookie(cookieVal)
	if err != nil {
		t.Fatalf("ParseDeveloperCookie returned error: %v", err)
	}
	if parsedID != devID {
		t.Fatalf("ParseDeveloperCookie ID mismatch: got %v, want %v", parsedID, devID)
	}
	if !utils.Verify([]byte(parsedID.String()), passwordHash, sig) {
		t.Fatal("Verify failed for signed developer cookie")
	}

	// Tampered cookie
	if utils.Verify([]byte(parsedID.String()), "different-password-hash", sig) {
		t.Fatal("Verify should fail with wrong password hash")
	}
}

func TestPublicEndpointDoesNotRequireAuth(t *testing.T) {
	r := router.NewServerMux(nil)

	// POST /api/mailboxes/custom without any auth header or cookie
	req := httptest.NewRequest(http.MethodPost, "/api/mailboxes/custom", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Since body is empty, it should fail with 400 Bad Request, NOT 401 Unauthorized
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request for unauthenticated public endpoint with empty body, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestDeveloperDashboardRequiresSession(t *testing.T) {
	r := router.NewServerMux(nil)

	// 1. Without session cookie -> 401
	req := httptest.NewRequest(http.MethodGet, "/api/dev/mailboxes", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized without session cookie, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 2. With invalid session cookie format -> 401
	req2 := httptest.NewRequest(http.MethodGet, "/api/dev/mailboxes", nil)
	req2.AddCookie(&http.Cookie{
		Name:  utils.DeveloperCookieName,
		Value: "malformed-cookie-value",
	})
	w2 := httptest.NewRecorder()

	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized with invalid cookie, got %d. Body: %s", w2.Code, w2.Body.String())
	}
}

func TestPublicAttachmentEndpointDoesNotRequireAuth(t *testing.T) {
	r := router.NewServerMux(nil)

	// Public use of the shared message endpoint must NOT require a developer
	// session. An invalid message id reaches 400 Bad Request before any DB
	// access — if auth were required we would see 401 instead.
	req := httptest.NewRequest(http.MethodGet, "/api/message/not-a-valid-id/attachment/0", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for unauthenticated public attachment request, got %d. Body: %s", w.Code, w.Body.String())
	}

	// A malformed developer cookie must also be ignored (optional session),
	// never turned into 401.
	req2 := httptest.NewRequest(http.MethodGet, "/api/message/not-a-valid-id/attachment/0", nil)
	req2.AddCookie(&http.Cookie{
		Name:  utils.DeveloperCookieName,
		Value: "malformed-cookie-value",
	})
	w2 := httptest.NewRecorder()

	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request with malformed developer cookie (optional session), got %d. Body: %s", w2.Code, w2.Body.String())
	}
}

func TestExternalV1APIRequiresAPIKey(t *testing.T) {
	r := router.NewServerMux(nil)

	// 1. Without Authorization header -> 401
	req := httptest.NewRequest(http.MethodGet, "/api/v1/mailboxes", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized without API key, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 2. With invalid Authorization header -> 401
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/mailboxes", nil)
	req2.Header.Set("Authorization", "InvalidFormat")
	w2 := httptest.NewRecorder()

	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized with invalid auth header, got %d. Body: %s", w2.Code, w2.Body.String())
	}
}

