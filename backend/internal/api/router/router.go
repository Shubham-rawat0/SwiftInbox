package router

import (
	"net/http"
	"time"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/api/handler"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/api/middleware"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
)

type Router struct {
	handler http.Handler
}

func NewServerMux(queries *postgres.Queries) *Router {

	mux := http.NewServeMux()

	messageAccessLimiter := middleware.NewRateLimiter(
		100, time.Minute,
		"Too many requests from this IP, please slow down",
		"MESSAGE")

	mailboxLimiter := middleware.NewRateLimiter(
		10, time.Hour,
		"Too many requests from this IP, please slow down",
		"MESSAGE",
	)

	generalLimiter := middleware.NewRateLimiter(
		200, time.Minute,
		"Too many requests from this IP, please try again later",
		"GENERAL",
	)

	apiMiddlewarehandler := middleware.NewApiMiddlewareHandler(queries)

	messageHandler := handler.NewMessageHandler(queries)
	mailboxHandler := handler.NewMailboxHandler(queries)
	apiHandler := handler.NewApiHandler(queries)

	mux.Handle("POST /api/mailboxes/custom", mailboxLimiter.Middleware(apiMiddlewarehandler.APIKey((http.HandlerFunc(mailboxHandler.CreateEmail)))))
	mux.Handle("POST /api/mailboxes", mailboxLimiter.Middleware(apiMiddlewarehandler.APIKey(http.HandlerFunc(mailboxHandler.CreateMailbox))))

	mux.Handle("POST /api/mailboxes/{address}/message", messageAccessLimiter.Middleware(apiMiddlewarehandler.APIKey(http.HandlerFunc(messageHandler.GetMessages))))
	mux.Handle("POST /api/message/{id}", messageAccessLimiter.Middleware(apiMiddlewarehandler.APIKey(http.HandlerFunc(messageHandler.GetMessage))))
	mux.Handle("POST /api/message/{id}/attachment/{index}", messageAccessLimiter.Middleware(apiMiddlewarehandler.APIKey(http.HandlerFunc(messageHandler.GetAttachment))))

	mux.Handle("POST /api/create", messageAccessLimiter.Middleware(http.HandlerFunc(apiHandler.AddApiKey)))
	mux.Handle("POST /api/dev/create", http.HandlerFunc(apiHandler.CreateDeveloper))

	mux.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.WriteJSON(w, 200, "healthy")
	}))

	handler := generalLimiter.Middleware(mux)

	handler = cors(handler)

	return &Router{
		handler: handler,
	}
}

func (h *Router) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.handler.ServeHTTP(w, r)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Origin", "127.0.0.1:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
