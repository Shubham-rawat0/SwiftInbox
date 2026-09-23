# SwiftInbox

SwiftInbox is a disposable mailbox service with a developer platform on top.

It provides public temporary inboxes, a developer dashboard for managing mailboxes, API keys and webhooks, and a metered external API at `/api/v1/*`.

## Architecture

```text
Browser
   │
   ▼
Next.js Frontend
   │
   │ HTTP / JSON
   ▼
Go Backend ───────────► PostgreSQL
   │
   ├── HTTP API
   ├── SMTP server
   ├── Webhook consumer ───► RabbitMQ ───► Developer webhooks
   └── Cleanup scheduler
```

Inbound mail arrives through SMTP, is stored in PostgreSQL, and is exposed through the HTTP API. Webhook events are processed asynchronously through RabbitMQ with retries and dead-letter handling.

## Stack

* **Backend:** Go, `net/http`, PostgreSQL, `pgx`, sqlc, goose
* **Messaging:** RabbitMQ, `amqp091-go`
* **SMTP:** `go-smtp`, `go-message`
* **Frontend:** Next.js, React, TypeScript, Tailwind CSS
* **API:** OpenAPI 3.1 + Swagger UI
* **Infrastructure:** Docker Compose

## Repository

```text
tempmail/
├── backend/       # Go API, SMTP, webhooks, cleanup and database
├── frontend/      # Next.js application and developer dashboard
└── docker-compose.yml
```

See [`backend/README.md`](backend/README.md) and [`frontend/README.md`](frontend/README.md) for implementation details.

## Quick Start

### Requirements

* Go 1.26+
* Node.js 20+
* PostgreSQL
* RabbitMQ

### Start infrastructure

```bash
docker compose up -d
```

### Backend

```bash
cd backend

# Configure backend/.env

goose -dir migrations postgres \
  "postgres://localhost:5432/tempmail?sslmode=disable" up

make run
```

The backend serves the HTTP API, SMTP server, webhook consumer and cleanup scheduler.

### Frontend

```bash
cd frontend

# Configure frontend/.env
npm install
npm run dev
```

The frontend runs at:

```text
http://localhost:3000
```

## API

SwiftInbox exposes three main API surfaces:

```text
/api/*       Public mailbox and authentication endpoints
/api/dev/*   Developer dashboard API
/api/v1/*    External developer API
```

The external developer API uses API keys:

```http
Authorization: Bearer dev_...
```

Requests to `/api/v1/*` are metered against the developer's usage quota.

### API Documentation

OpenAPI specification:

```text
/docs/openapi.json
```

Interactive Swagger UI:

```text
/docs/
```

## Webhooks

Developers can subscribe to events such as:

```text
email.received
email.deleted
mailbox.expired
```

Webhook deliveries are signed with HMAC and processed asynchronously through RabbitMQ with retry and dead-letter handling.

## Development

Backend tests:

```bash
cd backend
go test ./...
```

Frontend:

```bash
cd frontend
npm run lint
npm run build
```

## Notes

SwiftInbox uses public disposable mailboxes. Anyone who knows a mailbox address may be able to read its messages.

Do not use it for passwords, financial information, authentication credentials, or other sensitive data.
