# SwiftInbox Backend

Go backend for **SwiftInbox**, a disposable (temp) mailbox service. One binary owns the whole ingest
path: it serves the HTTP API, receives mail over SMTP, fans webhook events out through RabbitMQ, and
runs the expiry cleanup scheduler.

## Tech

| Concern | Choice |
| --- | --- |
| Language | Go 1.26 |
| HTTP | `net/http` `ServeMux` (method + path patterns), no framework |
| Database | PostgreSQL via `database/sql` + `pgx/v5` stdlib driver |
| Query layer | [sqlc](https://sqlc.dev) generated code in `internal/repository/postgres` |
| Migrations | goose-format SQL files in `migrations/` |
| SMTP | `github.com/emersion/go-smtp` |
| MIME parsing | `github.com/emersion/go-message` (wrapper in `internal/parser`) |
| Queue | RabbitMQ (`rabbitmq/amqp091-go`) with retry + dead-letter queues |
| Auth | HMAC-signed developer cookie, bcrypt password hashes, SHA-256 API key hashes |
| Rate limiting | `golang.org/x/time/rate` (per-IP) |
| API docs | Embedded OpenAPI 3.1 spec + Swagger UI (`internal/api/docs`) |

Module path: `github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend`

## Layout

```
cmd/api/main.go            entrypoint: DB, RabbitMQ consumer, SMTP server, HTTP server, scheduler
cmd/worker/worker.go       RabbitMQ broker setup (queues, retry queue, DLQ)
internal/api/server.go     http.Server wrapper with graceful shutdown
internal/api/router/       route table, CORS, rate limiters
internal/api/handler/      HTTP handlers (message, mailbox, developer, api key, webhook)
internal/api/middleware/   session cookie, API key, per-IP rate limiter
internal/api/docs/         embedded openapi.json + Swagger UI static assets
internal/service/          business logic (mailbox, api key, webhook, event validation)
internal/repository/       sqlc generated queries and models
internal/smtp/             SMTP backend: store inbound mail, publish email.received
internal/parser/           MIME -> text/html/attachments
internal/events/           RabbitMQ publisher/consumer, event fan-out, retry & DLQ
internal/cleanup/          expiry cleanup + jittered scheduler
internal/utils/            context values, crypto/HMAC, API key hashing, email helpers
migrations/                goose migrations
queries/                   sqlc query definitions
webhook/                   standalone webhook receiver used for local testing
test/                      parser + auth architecture tests
```

## Routes

All API responses are JSON; errors use `{"error": "<message>"}`. Routes are registered in
`internal/api/router/router.go`.

### 1. Public (no auth, rate-limited)

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/api/mailboxes/custom` | Create a mailbox from a `username` (10/hour/IP) |
| `POST` | `/api/mailboxes` | Create a mailbox from a full address, optional `expiresAt` |
| `POST` | `/api/mailboxes/{address}/message` | List messages in a mailbox (100/min/IP) |
| `POST` | `/api/message/{id}` | Fetch one message (parsed text/HTML/attachments) |
| `GET` | `/api/message/{id}/attachment/{index}` | Download an attachment |
| `POST` | `/api/dev/create` | Register a developer account |
| `POST` | `/api/dev/signin` | Sign in, sets the signed session cookie |
| `POST` | `/api/dev/signout` | Clear the session cookie |
| `GET` | `/health` | Liveness check, returns `"healthy"` |
| `GET` | `/docs/{path...}` | Swagger UI (`/docs/`) and `/docs/openapi.json` |

The message routes use `OptionalDeveloperSession`: a valid developer cookie is honoured (so private
developer mailboxes can be read), but the request is never rejected when the cookie is missing.
Everything is wrapped by a general 200/min/IP limiter plus the CORS handler (allows
`http://localhost:3000` with credentials).

### 2. Developer dashboard (`/api/dev/*`, requires session cookie)

| Method | Path |
| --- | --- |
| `GET` | `/api/dev/{id}` |
| `POST` | `/api/dev/mailboxes` |
| `POST` | `/api/dev/mailboxes/custom` |
| `GET` | `/api/dev/mailboxes` |
| `DELETE` | `/api/dev/mailboxes/{identifier}` |
| `GET` / `POST` | `/api/dev/keys` |
| `DELETE` | `/api/dev/keys/{id}` |
| `GET` / `POST` | `/api/dev/webhooks` |
| `GET` | `/api/dev/webhooks/dead-letters` |
| `POST` | `/api/dev/webhooks/dead-letters/{id}/seen` |
| `GET` / `PATCH` / `DELETE` | `/api/dev/webhooks/{id}` |
| `POST` | `/api/dev/webhooks/{id}/mailboxes` |
| `DELETE` | `/api/dev/webhooks/{id}/mailboxes/{mailboxID}` |
| `POST` | `/api/dev/webhooks/{id}/events/add` |
| `DELETE` | `/api/dev/webhooks/{id}/events/remove` |
| `POST` | `/api/dev/webhooks/{id}/test` |

### 3. External Developer API (`/api/v1/*`, requires `Authorization: Bearer dev_...`)

Every route passes through `RequireAPIKey` (SHA-256 hashes the key, rejects unknown/revoked keys with
401) and `WithUsage` (meters the request against `developer_usage` + `api_key_usage`). Exceeding the
quota returns `429` with `{"error":"usage quota exceeded"}`.

| Method | Path | Usage operation (credits) |
| --- | --- | --- |
| `POST` | `/api/v1/mailboxes` | `mailbox.create` (1) |
| `POST` | `/api/v1/mailboxes/custom` | `mailbox.create` (1) |
| `GET` | `/api/v1/mailboxes` | `mailbox.list` (1) |
| `GET` | `/api/v1/mailboxes/{id}` | `mailbox.list` (1) |
| `DELETE` | `/api/v1/mailboxes/{id}` | `mailbox.delete` (1) |
| `DELETE` | `/api/v1/mailboxes/address/{address}` | `mailbox.delete` (1) |
| `GET` | `/api/v1/mailboxes/{address}/messages` | `message.list` (3) |
| `POST` | `/api/v1/mailboxes/{address}/message` | `message.list` (3) |
| `GET` | `/api/v1/messages/{id}` | `message.get` (3) |
| `POST` | `/api/v1/message/{id}` | `message.get` (3) |
| `GET` | `/api/v1/messages/{id}/attachment/{index}` | `attachment.get` (2) |
| `GET` / `POST` | `/api/v1/webhooks` | `webhook.get` (1) / `webhook.create` (5) |
| `POST` | `/api/v1/webhooks/create` | `webhook.create` (5) |
| `GET` | `/api/v1/webhooks/dead-letters` | `webhook.dead_letters` (1) |
| `POST` | `/api/v1/webhooks/dead-letters/{id}/seen` | `webhook.dead_letter_seen` (1) |
| `GET` / `PATCH` / `DELETE` | `/api/v1/webhooks/{id}` | `webhook.get`/`webhook.update`/`webhook.delete` (1) |
| `POST` | `/api/v1/webhooks/{id}/mailboxes` | `webhook.create` (5) |
| `DELETE` | `/api/v1/webhooks/{id}/mailboxes/{mailboxID}` | `webhook.delete` (1) |
| `POST` | `/api/v1/webhooks/{id}/events/add` | `webhook.events.add` (1) |
| `DELETE` | `/api/v1/webhooks/{id}/events/remove` | `webhook.events.remove` (1) |
| `POST` | `/api/v1/webhooks/{id}/test` | `webhook.test` (2) |


## How it works

### Inbound mail (SMTP)

`internal/smtp/server.go` listens on `SMTP_PORT` (default `2525`) for the configured domain. On
`RCPT TO` it rejects any other domain, then upserts the mailbox with a 24-hour expiry. On `DATA` it
parses the MIME body (`internal/parser`) and stores one message row per recipient as raw bytes. Mail
delivered to a developer-owned mailbox expires in 30 days instead of 24 hours. After storing, an
`email.received` webhook event is published.

### Expiry cleanup

`internal/cleanup` deletes expired messages, expired mailboxes and long-revoked API keys. The
scheduler runs only when `CLEANUP_ENABLED=1`, starts after `CLEANUP_START_DELAY_MS` (default half the
interval, capped at 5 minutes) and re-schedules with ±10% jitter every `CLEANUP_INTERVAL_MS` (default
6h), shrinking to 30 minutes when something was deleted. Deletions emit `email.deleted` and
`mailbox.expired` webhook events.

### Webhooks

`internal/events` declares three durable queues:

- `webhook-events` – main work queue (8 workers, prefetch = workers)
- `webhook-events-retry` – TTL queue that dead-letters back into the main queue
- `webhook-events-dead-letter` – parked failed deliveries

Events currently emitted: `email.received`, `email.deleted`, `mailbox.expired`. Delivery is an HTTP
`POST` to the webhook URL with `Content-Type: application/json`, `X-Webhook-Event` and
`X-Webhook-Signature` (HMAC-SHA256 of the raw body, base64url, prefixed `sha256=`). Retries back off
10s → 30s → 2m → 10m (5 attempts total); after that the event is recorded in `webhook_dead_letters`
and published to the DLQ. Signing secrets are stored encrypted with `WEBHOOK_ENCRYPTION_KEY`.

## Configuration

`cmd/api/main.go` loads `.env` with godotenv; every value can also come from the environment. The
RabbitMQ URL is currently hardcoded to `amqp://guest:guest@localhost:5672/` in
`cmd/worker/worker.go`.

| Variable | Required | Default | Notes |
| --- | --- | --- | --- |
| `DB_URL` | yes | – | PostgreSQL connection string |
| `PORT` | yes | – | HTTP listen address, e.g. `:8080` |
| `SMTP_DOMAIN` / `MAIL_DOMAIN` | yes | `temp.mail.at` | Domain accepted by SMTP and appended to custom addresses |
| `SMTP_PORT` | no | `2525` | SMTP listener port |
| `API_KEY` | no | – | Shared/legacy key value |
| `CLEANUP_ENABLED` | no | – | Set to `1` to start the cleanup scheduler |
| `CLEANUP_INTERVAL_MS` | no | `21600000` (6h) | Cleanup interval |
| `CLEANUP_START_DELAY_MS` | no | min(interval/2, 5m) | Delay before the first cleanup run |
| `REVOKED_API_KEY_RETENTION_MS` | no | 7 days | Retention before revoked keys are deleted |
| `WEBHOOK_ENCRYPTION_KEY` | yes (webhooks) | – | Encrypts stored webhook signing secrets |
| `WEBHOOK_PORT` | no | `9090` | Only used by the `webhook/` test receiver |
| `WEBHOOK_SECRET` | no | – | Only used by the `webhook/` test receiver |

## Database

Migrations are goose-formatted (`-- +goose Up` / `-- +goose Down`) and `sqlc.yaml` points at the same
`migrations/` directory, so the generated query code and the schema stay in sync.

| Table | Purpose |
| --- | --- |
| `mailboxes` | `id`, unique `address`, `expires_at`, `created_by` (nullable developer) |
| `messages` | Raw MIME body, sender/subject, `mailbox_id` (cascade), `expires_at` |
| `developer` | Account, bcrypt `password_hash`, per-period quotas |
| `apikeys` | `key_hash` (SHA-256), `last_used_at`, `revoked_at` |
| `developer_usage` | Per developer/period counters per category |
| `api_key_usage` | Same counters per API key |
| `webhooks` | URL, encrypted secret, `events TEXT[]`, `is_active`, unique `(developer_id, url)` |
| `webhook_mailboxes` | Join table linking webhooks to mailboxes |
| `webhook_dead_letters` | Failed deliveries with `reason`, `attempts`, `seen` |

After editing `queries/*.sql` or adding a migration:

```bash
sqlc generate                                       # regenerate internal/repository/postgres
goose -dir migrations postgres "$DB_URL" up         # apply migrations
```

## Running

```bash
# PostgreSQL and RabbitMQ must be reachable.
goose -dir migrations postgres "$DB_URL" up

make run     # go build -o bin/fs ./cmd/api && ./bin/fs
make build   # build only
make test    # go test ./test
go test ./...  # full suite, including internal/api/docs
```

Optional local webhook receiver (accepts `POST /webhook`, verifies `WEBHOOK_SECRET`, logs payloads):

```bash
cd webhook && WEBHOOK_PORT=9090 go run .
```

## Tests

- `test/parser_test.go` – MIME parsing, attachments, inline attachments
- `test/auth_architecture_test.go` – session cookie helpers and route protection guarantees (public
  routes stay public, `/api/dev/*` needs a session, `/api/v1/*` needs an API key)
- `internal/api/docs/docs_test.go` – OpenAPI spec serving, Swagger UI index, static assets

