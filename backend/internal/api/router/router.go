package router

import (
	"net/http"
	"time"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/api/handler"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/api/middleware"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/utils"
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
	developerHandler := handler.NewDeveloperHandler(queries)
	webhookHandler := handler.NewWebhookHandler(queries)

	mux.Handle("POST /api/mailboxes/custom", mailboxLimiter.Middleware(utils.WithUsage("mailbox.create", apiMiddlewarehandler.APIKey(http.HandlerFunc(mailboxHandler.CreateEmail)))))
	mux.Handle("POST /api/mailboxes", mailboxLimiter.Middleware(utils.WithUsage("mailbox.create", apiMiddlewarehandler.APIKey(http.HandlerFunc(mailboxHandler.CreateMailbox)))))
	mux.Handle("DELETE /api/mailboxes/{address}", mailboxLimiter.Middleware(utils.WithUsage("mailbox.delete", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(mailboxHandler.DeleteMailbox)))))

	mux.Handle("POST /api/mailboxes/{address}/message", messageAccessLimiter.Middleware(utils.WithUsage("message.list", apiMiddlewarehandler.APIKey(http.HandlerFunc(messageHandler.GetMessages)))))
	mux.Handle("POST /api/message/{id}", messageAccessLimiter.Middleware(utils.WithUsage("message.get", apiMiddlewarehandler.APIKey(http.HandlerFunc(messageHandler.GetMessage)))))
	mux.Handle("POST /api/message/{id}/attachment/{index}", messageAccessLimiter.Middleware(utils.WithUsage("attachment.get", apiMiddlewarehandler.APIKey(http.HandlerFunc(messageHandler.GetAttachment)))))

	mux.Handle("POST /api/create", messageAccessLimiter.Middleware(http.HandlerFunc(apiHandler.AddApiKey)))
	mux.Handle("DELETE /api/revoke", messageAccessLimiter.Middleware(apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(apiHandler.RevokeApiKey))))
	mux.Handle("POST /api/dev/create", http.HandlerFunc(developerHandler.CreateDeveloper))
	mux.Handle("GET /api/dev/{id}", http.HandlerFunc(developerHandler.GetDeveloper))
	mux.Handle("GET /api/dev/{id}/keys", http.HandlerFunc(apiHandler.GetApiKeyUsage))

	mux.Handle("POST /api/webhooks/create", utils.WithUsage("webhook.create", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(webhookHandler.CreateWebhook))))
	mux.Handle("POST /api/webhook/create", utils.WithUsage("webhook.create", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(webhookHandler.CreateWebhook))))
	mux.Handle("GET /api/webhooks", utils.WithUsage("webhook.get", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(webhookHandler.ListWebhooks))))
	mux.Handle("GET /api/webhooks/{id}", utils.WithUsage("webhook.get", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(webhookHandler.GetWebhook))))
	mux.Handle("POST /api/webhooks/{id}/test", utils.WithUsage("webhook.test", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(webhookHandler.TestWebhook))))
	mux.Handle("DELETE /api/webhooks/{id}", utils.WithUsage("webhook.delete", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(webhookHandler.DeleteWebhook))))

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
