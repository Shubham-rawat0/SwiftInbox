# SwiftInbox Frontend

Next.js (App Router) client for **SwiftInbox**. It renders the public temp-mailbox experience, the
developer dashboard (mailboxes, API keys, webhooks) and an embedded Swagger UI for the external API.
Everything is client-rendered and talks directly to the Go backend over HTTP.

## Tech

| Concern | Choice |
| --- | --- |
| Framework | Next.js 16 (App Router) + React 19, TypeScript |
| Styling | Tailwind CSS v4, `tw-animate-css`, CSS variables in `app/globals.css` |
| UI kit | shadcn-style components in `components/ui` (built on `@base-ui/react`) |
| Icons | `lucide-react` |
| Theming | `next-themes` (system / light / dark) |
| Toasts | `sonner` |
| HTML email rendering | `dompurify` sanitising (`lib/sanitize.ts`) |

## Layout

```
app/                     routes (App Router)
components/ui/           shadcn-style primitives (button, input, sidebar, sheet, ...)
components/layout/       Header, Sidebar, Footer/RootFooter, DeveloperProfileContext
components/mailbox/      MailboxView, MessageDetailView, AttachmentView
components/developer/    DeveloperSessionGate, MailboxNavigation
components/docs/         DocsIframe (embeds backend Swagger UI)
hooks/                   use-developer-session, use-mobile
lib/api.ts               typed API client for the backend
lib/types.ts             response/request types mirroring the backend JSON
lib/sanitize.ts          DOMPurify config for message HTML
```

## Routes

| Path | Description |
| --- | --- |
| `/` | Create a mailbox from a username, or open an existing address |
| `/mailbox/[username]` | Public mailbox: message list, auto-refresh, copy address |
| `/mailbox/[username]/message/[messageId]` | Message detail (text/HTML, sanitised) |
| `/mailbox/[username]/message/[messageId]/attachment/[index]` | Attachment preview/download |
| `/developer` | Developer overview (profile + quotas) |
| `/developer/mailbox` | List, copy and delete developer-owned mailboxes |
| `/developer/mailbox/open` | Open an existing mailbox by address |
| `/developer/mailbox/[username]` | Developer mailbox view (session-authenticated fetch) |
| `/developer/mailbox/[username]/message/[messageId]` | Developer message detail |
| `/developer/mailbox/[username]/message/[messageId]/attachment/[index]` | Developer attachment view |
| `/developer/apikeys` | Create, list, view usage of and revoke API keys |
| `/developer/webhook` | Manage webhooks, events, linked mailboxes and dead-letter deliveries |
| `/docs` | Full-page OpenAPI 3.1 reference (iframe over the backend `/docs/`) |
| `/signin`, `/signup`, `/signout` | Developer authentication flows |

`app/developer/layout.tsx` wraps every dashboard page in `SidebarProvider` +
`DeveloperProfileProvider` + `DeveloperSessionGate`, so dashboard routes never render without a
resolved session.

## Data flow

- `lib/api.ts` is the single API client. Base URL comes from `NEXT_PUBLIC_API_BASE`
  (default `http://localhost:3001`).
- Requests use `credentials: 'include'` only for session-authenticated calls
  (`/api/dev/*`, and the `auth = true` variants of `fetchMessages` / `fetchMessage` /
  `fetchAttachment` that hit `/api/mailboxes/...`). Public calls omit credentials.
- Responses are cached in memory for 6 seconds and in-flight requests are de-duplicated by cache key,
  so a poll and a user-triggered refresh share one network call. `clearCache()` /
  `clearCacheForAddress()` invalidate after writes.
- Failures throw `ApiError` with the HTTP status, which pages use to distinguish "signed out" (401)
  from transient errors.
- Session state: the backend sets an HMAC-signed HttpOnly cookie on sign-in; the frontend also keeps
  `developer_id` in `localStorage` (used by the sidebar for UI) and dispatches a
  `developer-session-change` event. `useDeveloperSession()` treats a 401 as unauthenticated and any
  other failure as a retryable error.
- Email HTML is sanitised with DOMPurify before rendering; plain-text bodies are rendered as text.

## Environment

Create `frontend/.env`:

| Variable | Example | Notes |
| --- | --- | --- |
| `NEXT_PUBLIC_API_BASE` | `http://localhost:3001` | Backend origin (must match a CORS-allowed origin: the backend allows `http://localhost:3000`) |
| `NEXT_PUBLIC_MAIL_DOMAIN` | `temp.mail.at` | Domain appended to usernames in the UI |
| `API_KEY` | `dev_...` | Optional key used for manual/live API calls |

## Scripts

```bash
npm install
npm run dev     # next dev on http://localhost:3000
npm run build   # production build
npm run start   # serve the production build
npm run lint    # eslint
```

## Notes

- There are no Next.js route handlers or server actions: all backend calls happen in the browser, so
  the backend must be running and CORS-configured for the frontend origin.
- The `/docs` page embeds the Swagger UI served by the backend at `/docs/`, which is why the iframe
  requires the backend to be reachable.
- Mailboxes and messages are public by design; the UI warns users not to use them for sensitive data.
