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
		"MESSAGE",
	)

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

	sessionMiddleware := middleware.NewSessionMiddleware(queries)
	apiMiddlewarehandler := middleware.NewApiMiddlewareHandler(queries)

	messageHandler := handler.NewMessageHandler(queries)
	mailboxHandler := handler.NewMailboxHandler(queries)
	apiHandler := handler.NewApiHandler(queries)
	developerHandler := handler.NewDeveloperHandler(queries)
	webhookHandler := handler.NewWebhookHandler(queries)

	// 1. PUBLIC (No auth, rate-limited)
	mux.Handle("POST /api/mailboxes/custom", mailboxLimiter.Middleware(http.HandlerFunc(mailboxHandler.CreateEmail)))
	mux.Handle("POST /api/mailboxes", mailboxLimiter.Middleware(http.HandlerFunc(mailboxHandler.CreateMailbox)))
	mux.Handle("POST /api/mailboxes/{address}/message", sessionMiddleware.OptionalDeveloperSession(messageAccessLimiter.Middleware(http.HandlerFunc(messageHandler.GetMessages))))
	mux.Handle("POST /api/message/{id}", sessionMiddleware.OptionalDeveloperSession(messageAccessLimiter.Middleware(http.HandlerFunc(messageHandler.GetMessage))))
	mux.Handle("GET /api/message/{id}/attachment/{index}", sessionMiddleware.OptionalDeveloperSession(messageAccessLimiter.Middleware(http.HandlerFunc(messageHandler.GetAttachment))))

	// Developer registration & sign in/out (Public)
	mux.Handle("POST /api/dev/create", http.HandlerFunc(developerHandler.CreateDeveloper))
	mux.Handle("POST /api/dev/signin", http.HandlerFunc(developerHandler.SignIn))
	mux.Handle("POST /api/dev/signout", http.HandlerFunc(developerHandler.SignOut))

	// Health check
	mux.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.WriteJSON(w, 200, "healthy")
	}))

	// 2. DEVELOPER DASHBOARD (/api/dev/* with RequireDeveloperSession)
	mux.Handle("GET /api/dev/{id}", sessionMiddleware.RequireDeveloperSession(http.HandlerFunc(developerHandler.GetDeveloper)))

	// Developer Mailboxes
	mux.Handle("POST /api/dev/mailboxes", sessionMiddleware.RequireDeveloperSession(http.HandlerFunc(mailboxHandler.CreateDeveloperMailbox)))
	mux.Handle("POST /api/dev/mailboxes/custom", sessionMiddleware.RequireDeveloperSession(http.HandlerFunc(mailboxHandler.CreateDeveloperEmail)))
	mux.Handle("GET /api/dev/mailboxes", sessionMiddleware.RequireDeveloperSession(http.HandlerFunc(mailboxHandler.ListDeveloperMailboxes)))
	mux.Handle("DELETE /api/dev/mailboxes/{identifier}", sessionMiddleware.RequireDeveloperSession(http.HandlerFunc(mailboxHandler.DeleteDeveloperMailbox)))

	// Developer API Key Management
	mux.Handle("GET /api/dev/keys", sessionMiddleware.RequireDeveloperSession(http.HandlerFunc(apiHandler.GetApiKeyUsage)))
	mux.Handle("POST /api/dev/keys", sessionMiddleware.RequireDeveloperSession(http.HandlerFunc(apiHandler.AddApiKey)))

	mux.Handle("DELETE /api/dev/keys/{id}", sessionMiddleware.RequireDeveloperSession(http.HandlerFunc(apiHandler.RevokeDeveloperApiKey)))

	// Developer Webhook Management
	mux.Handle("GET /api/dev/webhooks", sessionMiddleware.RequireDeveloperSession(http.HandlerFunc(webhookHandler.ListWebhooks)))
	mux.Handle("POST /api/dev/webhooks", sessionMiddleware.RequireDeveloperSession(http.HandlerFunc(webhookHandler.CreateWebhook)))
	mux.Handle("GET /api/dev/webhooks/dead-letters", sessionMiddleware.RequireDeveloperSession(http.HandlerFunc(webhookHandler.GetDeadLetters)))
	mux.Handle("GET /api/dev/webhooks/{id}", sessionMiddleware.RequireDeveloperSession(http.HandlerFunc(webhookHandler.GetWebhook)))
	mux.Handle("DELETE /api/dev/webhooks/{id}", sessionMiddleware.RequireDeveloperSession(http.HandlerFunc(webhookHandler.DeleteWebhook)))
	mux.Handle("POST /api/dev/webhooks/{id}/mailboxes", sessionMiddleware.RequireDeveloperSession(http.HandlerFunc(webhookHandler.AddMailbox)))
	mux.Handle("DELETE /api/dev/webhooks/{id}/mailboxes/{mailboxID}", sessionMiddleware.RequireDeveloperSession(http.HandlerFunc(webhookHandler.RemoveMailbox)))
	mux.Handle("POST /api/dev/webhooks/{id}/events/add", sessionMiddleware.RequireDeveloperSession(http.HandlerFunc(webhookHandler.AddEvents)))
	mux.Handle("DELETE /api/dev/webhooks/{id}/events/remove", sessionMiddleware.RequireDeveloperSession(http.HandlerFunc(webhookHandler.RemoveEvents)))
	mux.Handle("POST /api/dev/webhooks/{id}/test", sessionMiddleware.RequireDeveloperSession(http.HandlerFunc(webhookHandler.TestWebhook)))

	// 3. EXTERNAL DEVELOPER API (/api/v1/* with RequireAPIKey + WithUsage)
	mux.Handle("POST /api/v1/mailboxes", utils.WithUsage("mailbox.create", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(mailboxHandler.CreateAPIMailbox))))
	mux.Handle("POST /api/v1/mailboxes/custom", utils.WithUsage("mailbox.create", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(mailboxHandler.CreateCustomAPIMailbox))))
	mux.Handle("GET /api/v1/mailboxes", utils.WithUsage("mailbox.list", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(mailboxHandler.ListAPIMailboxes))))
	mux.Handle("GET /api/v1/mailboxes/{id}", utils.WithUsage("mailbox.list", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(mailboxHandler.GetAPIMailbox))))
	mux.Handle("DELETE /api/v1/mailboxes/{id}", utils.WithUsage("mailbox.delete", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(mailboxHandler.DeleteAPIMailbox))))
	mux.Handle("DELETE /api/v1/mailboxes/address/{address}", utils.WithUsage("mailbox.delete", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(mailboxHandler.DeleteAPIMailbox))))

	// Messages & Attachments
	mux.Handle("GET /api/v1/mailboxes/{address}/messages", utils.WithUsage("message.list", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(messageHandler.GetMessages))))
	mux.Handle("POST /api/v1/mailboxes/{address}/message", utils.WithUsage("message.list", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(messageHandler.GetMessages))))
	mux.Handle("GET /api/v1/messages/{id}", utils.WithUsage("message.get", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(messageHandler.GetMessage))))
	mux.Handle("POST /api/v1/message/{id}", utils.WithUsage("message.get", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(messageHandler.GetMessage))))
	mux.Handle("GET /api/v1/messages/{id}/attachment/{index}", utils.WithUsage("attachment.get", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(messageHandler.GetAttachment))))

	// Webhooks
	mux.Handle("GET /api/v1/webhooks", utils.WithUsage("webhook.get", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(webhookHandler.ListWebhooks))))
	mux.Handle("POST /api/v1/webhooks", utils.WithUsage("webhook.create", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(webhookHandler.CreateWebhook))))
	mux.Handle("POST /api/v1/webhooks/create", utils.WithUsage("webhook.create", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(webhookHandler.CreateWebhook))))
	mux.Handle("GET /api/v1/webhooks/dead-letters", utils.WithUsage("webhook.dead_letters", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(webhookHandler.GetDeadLetters))))
	mux.Handle("GET /api/v1/webhooks/{id}", utils.WithUsage("webhook.get", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(webhookHandler.GetWebhook))))
	mux.Handle("DELETE /api/v1/webhooks/{id}", utils.WithUsage("webhook.delete", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(webhookHandler.DeleteWebhook))))
	mux.Handle("POST /api/v1/webhooks/{id}/mailboxes", utils.WithUsage("webhook.create", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(webhookHandler.AddMailbox))))
	mux.Handle("DELETE /api/v1/webhooks/{id}/mailboxes/{mailboxID}", utils.WithUsage("webhook.delete", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(webhookHandler.RemoveMailbox))))
	mux.Handle("POST /api/v1/webhooks/{id}/events/add", utils.WithUsage("webhook.events.add", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(webhookHandler.AddEvents))))
	mux.Handle("DELETE /api/v1/webhooks/{id}/events/remove", utils.WithUsage("webhook.events.remove", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(webhookHandler.RemoveEvents))))
	mux.Handle("POST /api/v1/webhooks/{id}/test", utils.WithUsage("webhook.test", apiMiddlewarehandler.RequireAPIKey(http.HandlerFunc(webhookHandler.TestWebhook))))

	handler := generalLimiter.Middleware(mux)
	handler = cors(handler)

	return &Router{
		handler: handler,
	}
}

func (h *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:3000" || origin == "http://127.0.0.1:3000" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Add("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
