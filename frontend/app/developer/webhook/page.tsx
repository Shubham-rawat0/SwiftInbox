"use client"

import {
  addWebhookEvents,
  createWebhook,
  deleteWebhook,
  isApiError,
  linkWebhookMailbox,
  listDeveloperMailboxes,
  listWebhookDeadLetters,
  listWebhooks,
  markWebhookDeadLetterSeen,
  removeWebhookEvents,
  testWebhook,
  unlinkWebhookMailbox,
  updateWebhook,
} from "@/lib/api"
import type {
  MailboxResponse,
  WebhookDeadLetterResponse,
  WebhookResponse,
} from "@/lib/types"
import {
  AlertTriangle,
  Check,
  ChevronRight,
  Copy,
  ExternalLink,
  Inbox,
  Link2,
  Loader2,
  Plus,
  Send,
  Trash2,
  Webhook,
  X,
} from "lucide-react"
import Link from "next/link"
import { useCallback, useEffect, useMemo, useState } from "react"
import { toast } from "sonner"

const EVENT_OPTIONS = [
  {
    key: "email.received",
    label: "Email received",
    description: "A message was delivered to a linked mailbox.",
  },
  {
    key: "email.deleted",
    label: "Email deleted",
    description: "A message in a linked mailbox was deleted or expired.",
  },
  {
    key: "mailbox.expired",
    label: "Mailbox expired",
    description: "A linked mailbox reached its expiry and was cleaned up.",
  },
] as const

export default function WebhooksPage() {
  return <WebhooksManager />
}

function normalizeWebhook(webhook: WebhookResponse): WebhookResponse {
  return {
    ...webhook,
    name: webhook.name ?? "",
    mailbox_ids: webhook.mailbox_ids ?? [],
    events: webhook.events ?? [],
  }
}

function WebhooksManager() {
  const [webhooks, setWebhooks] = useState<WebhookResponse[] | null>(null)
  const [mailboxes, setMailboxes] = useState<MailboxResponse[] | null>(null)
  const [deadLetters, setDeadLetters] = useState<WebhookDeadLetterResponse[] | null>(null)
  const [loading, setLoading] = useState(true)
  const [signedOut, setSignedOut] = useState(false)
  const [failed, setFailed] = useState(false)
  const [reloadKey, setReloadKey] = useState(0)
  const [showCreate, setShowCreate] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<WebhookResponse | null>(null)

  useEffect(() => {
    let isActive = true

    async function loadDeadLetters() {
      try {
        const result = await listWebhookDeadLetters()
        if (isActive) setDeadLetters(result)
      } catch {
        if (isActive) setDeadLetters([])
      }
    }

    async function load() {
      try {
        const [webhooksResult, mailboxesResult] = await Promise.all([
          listWebhooks(),
          listDeveloperMailboxes(),
        ])

        if (!isActive) return

        setWebhooks(webhooksResult.map(normalizeWebhook))
        setMailboxes(mailboxesResult)
        setFailed(false)
      } catch (error) {
        if (!isActive) return

        setWebhooks([])
        setMailboxes([])
        setDeadLetters([])

        if (isApiError(error) && error.status === 401) {
          setSignedOut(true)
        } else {
          setFailed(true)
        }
      } finally {
        if (isActive) {
          setLoading(false)
        }
      }
    }

    void load()
    void loadDeadLetters()

    return () => {
      isActive = false
    }
  }, [reloadKey])

  const onCreated = (created: WebhookResponse) => {
    const normalized = normalizeWebhook(created)
    setShowCreate(false)
    setWebhooks((current) =>
      current === null ? current : [normalized, ...current.filter((w) => w.id !== normalized.id)]
    )
  }

  const onUpdated = (updated: WebhookResponse) => {
    const normalized = normalizeWebhook(updated)
    setWebhooks((current) =>
      current === null
        ? current
        : current.map((webhook) => (webhook.id === normalized.id ? normalized : webhook))
    )
  }

  const onDelete = (id: string) => {
    setWebhooks((current) =>
      current === null ? current : current.filter((webhook) => webhook.id !== id)
    )
  }

  const onLink = useCallback((webhookId: string, mailboxId: string) => {
    setWebhooks((current) =>
      current === null
        ? current
        : current.map((webhook) =>
            webhook.id === webhookId && !webhook.mailbox_ids.includes(mailboxId)
              ? { ...webhook, mailbox_ids: [...webhook.mailbox_ids, mailboxId] }
              : webhook
          )
    )
  }, [])

  const onUnlink = useCallback((webhookId: string, mailboxId: string) => {
    setWebhooks((current) =>
      current === null
        ? current
        : current.map((webhook) =>
            webhook.id === webhookId
              ? {
                  ...webhook,
                  mailbox_ids: webhook.mailbox_ids.filter((id) => id !== mailboxId),
                }
              : webhook
          )
    )
  }, [])

  const showContent = !loading && !signedOut && !failed

  return (
    <div className="min-h-[calc(100svh-4rem)] bg-[#f8f8f6] text-[#171717] dark:bg-[#0b0c0c] dark:text-white">
      <main className="mx-auto w-full max-w-4xl px-5 py-6 sm:px-8 sm:py-8">
        <div className="animate-in fade-in duration-300 ease-out">
          <div className="mb-6 flex flex-wrap items-start justify-between gap-4">
            <div className="max-w-xl">
              <h1 className="font-heading text-2xl font-semibold tracking-[-0.035em] sm:text-3xl">
                Webhooks
              </h1>

              <p className="mt-1.5 text-[14px] leading-5 text-black/45 dark:text-white/40">
                Receive realtime HTTP callbacks when something happens on your
                mailboxes. Delivered with an{" "}
                <code className="rounded bg-black/[0.05] px-1 font-mono text-[12px] dark:bg-white/[0.07]">
                  X-Webhook-Signature
                </code>{" "}
                header you can use to verify requests.
              </p>
            </div>

            {showContent && (
              <button
                type="button"
                onClick={() => setShowCreate(true)}
                className="inline-flex h-10 shrink-0 items-center gap-1.5 rounded-xl bg-black px-4 text-[13px] font-semibold text-white transition hover:bg-black/80 active:scale-[0.99] dark:bg-white dark:text-black dark:hover:bg-white/85"
              >
                <Plus className="size-4" />
                Create webhook
              </button>
            )}
          </div>

          {loading ? (
            <div className="flex min-h-[320px] items-center justify-center">
              <div className="flex items-center gap-2.5 text-sm text-muted-foreground">
                <div className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                Loading webhooks
              </div>
            </div>
          ) : signedOut ? (
            <div className="flex min-h-[320px] items-center justify-center px-6">
              <div className="text-center">
                <p className="text-sm font-medium">Sign in to view your webhooks</p>

                <p className="mt-1 text-sm text-muted-foreground">
                  Your developer session is required to continue.
                </p>

                <Link
                  href="/signin"
                  className="mt-6 inline-flex h-10 items-center justify-center rounded-xl bg-black px-5 text-sm font-semibold text-white transition hover:bg-black/80 dark:bg-white dark:text-black dark:hover:bg-white/85"
                >
                  Sign in
                </Link>
              </div>
            </div>
          ) : failed ? (
            <div className="flex min-h-[320px] items-center justify-center px-6">
              <div className="text-center">
                <p className="text-sm font-medium">Couldn&apos;t load your webhooks</p>

                <p className="mt-1 text-sm text-muted-foreground">
                  A temporary problem occurred while reaching the server.
                </p>

                <button
                  type="button"
                  onClick={() => {
                    setLoading(true)
                    setFailed(false)
                    setReloadKey((value) => value + 1)
                  }}
                  className="mt-6 inline-flex h-10 items-center justify-center rounded-xl border border-black/10 bg-white px-5 text-sm font-semibold text-black transition hover:bg-black/[0.03] dark:border-white/10 dark:bg-white/[0.04] dark:text-white dark:hover:bg-white/[0.07]"
                >
                  Try again
                </button>
              </div>
            </div>
          ) : (
            <>
              <WebhooksList
                webhooks={webhooks ?? []}
                mailboxes={mailboxes ?? []}
                onCreate={() => setShowCreate(true)}
                onDelete={setDeleteTarget}
                onUpdate={onUpdated}
                onLink={onLink}
                onUnlink={onUnlink}
              />

              {deadLetters !== null && deadLetters.length > 0 && (
                <DeadLetters
                  deadLetters={deadLetters}
                  onMarkSeen={(id) =>
                    setDeadLetters((current) =>
                      current === null
                        ? current
                        : current.map((deadLetter) =>
                            deadLetter.id === id
                              ? { ...deadLetter, seen: true }
                              : deadLetter
                          )
                    )
                  }
                />
              )}
            </>
          )}
        </div>
      </main>

      {showCreate && (
        <CreateWebhookDialog
          mailboxes={mailboxes ?? []}
          onClose={() => setShowCreate(false)}
          onCreated={onCreated}
        />
      )}

      {deleteTarget && (
        <DeleteWebhookDialog
          key={deleteTarget.id}
          target={deleteTarget}
          onCancel={() => setDeleteTarget(null)}
          onDeleted={() => {
            onDelete(deleteTarget.id)
            setDeleteTarget(null)
          }}
        />
      )}
    </div>
  )
}

function WebhooksList({
  webhooks,
  mailboxes,
  onCreate,
  onDelete,
  onUpdate,
  onLink,
  onUnlink,
}: {
  webhooks: WebhookResponse[]
  mailboxes: MailboxResponse[]
  onCreate: () => void
  onDelete: (webhook: WebhookResponse) => void
  onUpdate: (webhook: WebhookResponse) => void
  onLink: (webhookId: string, mailboxId: string) => void
  onUnlink: (webhookId: string, mailboxId: string) => void
}) {
  const [expandedId, setExpandedId] = useState<string | null>(null)

  if (webhooks.length === 0) {
    return (
      <section className="overflow-hidden rounded-2xl border border-black/[0.09] bg-white shadow-[0_12px_40px_rgba(0,0,0,0.05)] dark:border-white/[0.08] dark:bg-[#111313] dark:shadow-black/20">
        <div className="flex min-h-[260px] flex-col items-center justify-center gap-3 px-6 text-center">
          <span className="flex size-12 items-center justify-center rounded-full bg-black/[0.04] dark:bg-white/[0.05]">
            <Webhook className="size-5 text-black/35 dark:text-white/30" />
          </span>

          <div>
            <p className="text-sm font-medium">No webhooks yet</p>
            <p className="mt-1 text-[13px] text-black/40 dark:text-white/35">
              Create a webhook to receive callbacks for your mailboxes.
            </p>
          </div>

          <button
            type="button"
            onClick={onCreate}
            className="mt-2 inline-flex h-9 items-center gap-1.5 rounded-lg bg-black px-3 text-[13px] font-semibold text-white transition hover:bg-black/80 active:scale-[0.99] dark:bg-white dark:text-black dark:hover:bg-white/85"
          >
            <Plus className="size-4" />
            Create webhook
          </button>
        </div>
      </section>
    )
  }

  return (
    <section className="overflow-hidden rounded-2xl border border-black/[0.09] bg-white shadow-[0_12px_40px_rgba(0,0,0,0.05)] dark:border-white/[0.08] dark:bg-[#111313] dark:shadow-black/20">
      <div className="divide-y divide-black/[0.06] dark:divide-white/[0.06]">
        {webhooks.map((webhook) => {
          return (
            <WebhookRow
              key={webhook.id}
              webhook={webhook}
              mailboxes={mailboxes}
              expanded={expandedId === webhook.id}
              onToggle={() =>
                setExpandedId((current) => (current === webhook.id ? null : webhook.id))
              }
              onDelete={() => onDelete(webhook)}
              onUpdate={onUpdate}
              onLink={onLink}
              onUnlink={onUnlink}
            />
          )
        })}
      </div>
    </section>
  )
}

function WebhookRow({
  webhook,
  mailboxes,
  expanded,
  onToggle,
  onDelete,
  onUpdate,
  onLink,
  onUnlink,
}: {
  webhook: WebhookResponse
  mailboxes: MailboxResponse[]
  expanded: boolean
  onToggle: () => void
  onDelete: () => void
  onUpdate: (webhook: WebhookResponse) => void
  onLink: (webhookId: string, mailboxId: string) => void
  onUnlink: (webhookId: string, mailboxId: string) => void
}) {
  const linkedIds = useMemo(() => webhook.mailbox_ids ?? [], [webhook.mailbox_ids])
  const [copied, setCopied] = useState(false)
  const [savingEvents, setSavingEvents] = useState(false)
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState<{ success: boolean; statusCode: number } | null>(null)
  const [addingMailbox, setAddingMailbox] = useState(false)
  const [selectedMailboxId, setSelectedMailboxId] = useState("")
  const [removingMailboxId, setRemovingMailboxId] = useState<string | null>(null)
  const [togglingActive, setTogglingActive] = useState(false)

  const subscribed = new Set(webhook.events)
  const availableMailboxes = useMemo(
    () => mailboxes.filter((mailbox) => !linkedIds.includes(mailbox.id)),
    [mailboxes, linkedIds]
  )
  const mailboxLookup = useMemo(
    () => new Map(mailboxes.map((mailbox) => [mailbox.id, mailbox.address])),
    [mailboxes]
  )

  const copyUrl = async () => {
    try {
      await navigator.clipboard.writeText(webhook.url)
      setCopied(true)
      toast.success("Webhook URL copied")
      setTimeout(() => setCopied(false), 1800)
    } catch {
      toast.error("Could not copy URL")
    }
  }

  const toggleActive = async () => {
    if (togglingActive) return

    setTogglingActive(true)
    try {
      const updated = await updateWebhook(webhook.id, { is_active: !webhook.is_active })
      onUpdate(updated)
      toast.success(updated.is_active ? "Webhook activated" : "Webhook paused")
    } catch (error) {
      toast.error("Could not update webhook", {
        description: error instanceof Error ? error.message : "Something went wrong.",
      })
    } finally {
      setTogglingActive(false)
    }
  }

  const toggleEvent = async (eventKey: string) => {
    if (savingEvents) return

    const willSubscribe = !subscribed.has(eventKey)
    setSavingEvents(true)

    try {
      const updated = willSubscribe
        ? await addWebhookEvents(webhook.id, [eventKey])
        : await removeWebhookEvents(webhook.id, [eventKey])
      onUpdate(updated)
      toast.success(willSubscribe ? "Event subscribed" : "Event removed")
    } catch (error) {
      toast.error("Could not update events", {
        description: error instanceof Error ? error.message : "Something went wrong.",
      })
    } finally {
      setSavingEvents(false)
    }
  }

  const runTest = async () => {
    setTesting(true)
    setTestResult(null)

    try {
      const result = await testWebhook(webhook.id)
      setTestResult(result)
      toast.success(
        result.success ? `Delivery OK (${result.statusCode})` : `Delivery failed (${result.statusCode})`
      )
    } catch (error) {
      toast.error("Test delivery failed", {
        description: error instanceof Error ? error.message : "Something went wrong.",
      })
    } finally {
      setTesting(false)
    }
  }

  const addMailbox = () => {
    if (!selectedMailboxId || addingMailbox) return

    const mailboxId = selectedMailboxId
    setAddingMailbox(true)
    setSelectedMailboxId("")

    linkWebhookMailbox(webhook.id, mailboxId)
      .then(() => {
        toast.success("Mailbox linked to webhook")
        onLink(webhook.id, mailboxId)
      })
      .catch((error: unknown) => {
        toast.error("Could not link mailbox", {
          description: error instanceof Error ? error.message : "Something went wrong.",
        })
      })
      .finally(() => setAddingMailbox(false))
  }

  const removeMailbox = (mailboxId: string) => {
    if (removingMailboxId) return

    setRemovingMailboxId(mailboxId)
    unlinkWebhookMailbox(webhook.id, mailboxId)
      .then(() => {
        toast.success("Mailbox unlinked from webhook")
        onUnlink(webhook.id, mailboxId)
      })
      .catch((error: unknown) => {
        toast.error("Could not unlink mailbox", {
          description: error instanceof Error ? error.message : "Something went wrong.",
        })
      })
      .finally(() => setRemovingMailboxId(null))
  }

  return (
    <div className={expanded ? "bg-black/[0.015] dark:bg-white/[0.015]" : ""}>
      <div className="flex items-center">
        <button
          type="button"
          onClick={onToggle}
          aria-expanded={expanded}
          className="grid min-w-0 flex-1 grid-cols-[16px_minmax(0,1fr)_auto] items-center gap-4 px-5 py-4 text-left transition-colors hover:bg-black/[0.02] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-black/20 dark:hover:bg-white/[0.03] dark:focus-visible:ring-white/25"
        >
          <ChevronRight
            aria-hidden="true"
            className={`size-4 text-black/35 transition-transform duration-200 dark:text-white/30 ${
              expanded ? "rotate-90" : ""
            }`}
          />

          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <p className="truncate text-[13px] font-medium text-black/75 dark:text-white/75">
                {webhook.name || webhook.url}
              </p>

              {!webhook.is_active && (
                <span className="shrink-0 rounded-md bg-amber-500/10 px-1.5 py-0.5 text-[11px] font-medium text-amber-700 dark:text-amber-300">
                  Inactive
                </span>
              )}
            </div>

            <p className="mt-0.5 truncate font-mono text-[11px] text-black/35 dark:text-white/30">
              {webhook.url}
            </p>

            <div className="mt-1.5 flex flex-wrap items-center gap-1.5">
              {webhook.events.length === 0 ? (
                <span className="text-[11px] text-black/30 dark:text-white/25">
                  No events subscribed
                </span>
              ) : (
                webhook.events.map((event) => (
                  <span
                    key={event}
                    className="rounded-md bg-black/[0.04] px-1.5 py-0.5 font-mono text-[10px] font-medium text-black/50 dark:bg-white/[0.06] dark:text-white/45"
                  >
                    {event}
                  </span>
                ))
              )}
            </div>
          </div>
        </button>

        <div className="flex shrink-0 items-center gap-1 pr-3">
          <a
            href={webhook.url}
            target="_blank"
            rel="noopener noreferrer"
            aria-label={`Open ${webhook.url}`}
            title="Open webhook URL"
            className="flex size-8 items-center justify-center rounded-lg text-black/30 transition-colors hover:bg-black/[0.05] hover:text-black/70 dark:text-white/30 dark:hover:bg-white/[0.07] dark:hover:text-white/70"
          >
            <ExternalLink className="size-4" />
          </a>

          <button
            type="button"
            aria-label="Delete webhook"
            title="Delete webhook"
            onClick={onDelete}
            className="flex size-8 items-center justify-center rounded-lg text-black/30 transition-colors hover:bg-red-500/10 hover:text-red-500 dark:text-white/30 dark:hover:text-red-500"
          >
            <Trash2 className="size-4" />
          </button>
        </div>
      </div>

      {expanded && (
        <div className="animate-in fade-in slide-in-from-top-1 border-t border-black/[0.06] px-5 pb-6 pt-5 duration-200 dark:border-white/[0.06] sm:pl-[52px]">
          <div className="space-y-7">
            {/* Name */}
            <div>
              <p className="text-sm font-semibold">Name</p>

              <WebhookNameEditor
                key={webhook.name}
                webhook={webhook}
                onUpdate={onUpdate}
              />
            </div>

            {/* Active */}
            <div className="flex items-center justify-between gap-4">
              <div>
                <p className="text-sm font-semibold">Active</p>
                <p className="mt-0.5 text-xs text-black/40 dark:text-white/35">
                  {webhook.is_active
                    ? "Events are being delivered to this endpoint."
                    : "Delivery is paused; events won’t be sent."}
                </p>
              </div>

              <button
                type="button"
                role="switch"
                aria-checked={webhook.is_active}
                aria-label="Toggle webhook active"
                onClick={() => void toggleActive()}
                disabled={togglingActive}
                className={`relative h-6 w-11 shrink-0 rounded-full transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${
                  webhook.is_active
                    ? "bg-black dark:bg-white"
                    : "bg-black/15 dark:bg-white/15"
                }`}
              >
                <span
                  aria-hidden="true"
                  className="absolute size-5 rounded-full bg-white shadow-md transition-transform duration-200 dark:bg-black"
                  style={{
                    top: 2,
                    left: 2,
                    transform: webhook.is_active ? "translateX(20px)" : "translateX(0)",
                  }}
                />
              </button>
            </div>

            {/* URL */}
            <div>
              <div className="flex items-center gap-2">
                <p className="text-sm font-semibold">Endpoint</p>
                <button
                  type="button"
                  aria-label="Copy webhook URL"
                  title="Copy webhook URL"
                  onClick={copyUrl}
                  className="flex size-7 items-center justify-center rounded-md text-black/30 transition-colors hover:bg-black/[0.05] hover:text-black/70 dark:text-white/30 dark:hover:bg-white/[0.07] dark:hover:text-white/70"
                >
                  {copied ? (
                    <Check className="size-3.5 text-emerald-600" />
                  ) : (
                    <Copy className="size-3.5" />
                  )}
                </button>
              </div>
              <code className="mt-2 block truncate rounded-xl border border-black/[0.08] bg-[#fafaf9] px-3.5 py-2.5 font-mono text-[12px] text-black/70 dark:border-white/[0.08] dark:bg-[#0b0c0c] dark:text-white/70">
                {webhook.url}
              </code>
            </div>

            {/* Events */}
            <div>
              <p className="text-sm font-semibold">Events</p>
              <p className="mt-0.5 text-xs text-black/40 dark:text-white/35">
                Toggle the delivery events this endpoint will receive.
              </p>

              <div
                className={`mt-3 grid gap-2 transition-opacity sm:grid-cols-3 ${
                  savingEvents ? "pointer-events-none opacity-50" : ""
                }`}
              >
                {EVENT_OPTIONS.map((event) => {
                  const isOn = subscribed.has(event.key)

                  return (
                    <button
                      key={event.key}
                      type="button"
                      role="checkbox"
                      aria-checked={isOn}
                      onClick={() => toggleEvent(event.key)}
                      className={`flex flex-col gap-1 rounded-xl border p-3 text-left transition ${
                        isOn
                          ? "border-black/[0.16] bg-black/[0.03] dark:border-white/[0.18] dark:bg-white/[0.05]"
                          : "border-black/[0.08] hover:border-black/[0.14] dark:border-white/[0.08] dark:hover:border-white/[0.15]"
                      }`}
                    >
                      <span className="flex items-center justify-between gap-2">
                        <code className="font-mono text-[11px] font-medium text-black/70 dark:text-white/70">
                          {event.key}
                        </code>
                        <span
                          className={`flex size-4 shrink-0 items-center justify-center rounded-full transition ${
                            isOn
                              ? "bg-black text-white dark:bg-white dark:text-black"
                              : "border border-black/20 dark:border-white/20"
                          }`}
                        >
                          {isOn && <Check className="size-2.5" />}
                        </span>
                      </span>
                      <span className="text-[11px] leading-4 text-black/40 dark:text-white/35">
                        {event.description}
                      </span>
                    </button>
                  )
                })}
              </div>
            </div>

            {/* Mailboxes */}
            <div>
              <p className="text-sm font-semibold">Linked mailboxes</p>
              <p className="mt-0.5 text-xs text-black/40 dark:text-white/35">
                Events are only delivered for addresses linked to this webhook.
              </p>

              <div className="mt-3 flex flex-wrap items-center gap-2">
                <select
                  aria-label="Mailbox to link"
                  value={selectedMailboxId}
                  onChange={(event) => setSelectedMailboxId(event.target.value)}
                  disabled={availableMailboxes.length === 0}
                  className="h-9 min-w-0 rounded-lg border border-black/[0.10] bg-[#fafaf9] px-2.5 text-[13px] font-medium outline-none transition focus:border-black/25 focus:ring-4 focus:ring-black/[0.035] disabled:cursor-not-allowed disabled:opacity-40 dark:border-white/[0.10] dark:bg-[#0b0c0c] dark:focus:border-white/25 dark:focus:ring-white/[0.035]"
                >
                  <option value="">
                    {availableMailboxes.length === 0
                      ? "No unlinked mailboxes"
                      : "Select a mailbox…"}
                  </option>
                  {availableMailboxes.map((mailbox) => (
                    <option key={mailbox.id} value={mailbox.id}>
                      {mailbox.address}
                    </option>
                  ))}
                </select>

                <button
                  type="button"
                  onClick={addMailbox}
                  disabled={!selectedMailboxId || addingMailbox}
                  className="inline-flex h-9 items-center gap-1.5 rounded-lg bg-black px-3 text-[13px] font-semibold text-white transition hover:bg-black/80 active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-30 dark:bg-white dark:text-black dark:hover:bg-white/85"
                >
                  <Link2 className="size-3.5" />
                  Link
                </button>
              </div>

              {linkedIds.length > 0 ? (
                <ul className="mt-4 space-y-2">
                  {linkedIds.map((mailboxId) => (
                    <li
                      key={mailboxId}
                      className="flex items-center justify-between gap-3 rounded-xl border border-black/[0.07] px-3 py-2 dark:border-white/[0.07]"
                    >
                      <span className="truncate font-mono text-[12px] text-black/70 dark:text-white/70">
                        {mailboxLookup.get(mailboxId) ?? mailboxId}
                      </span>

                      <button
                        type="button"
                        aria-label={`Unlink ${mailboxLookup.get(mailboxId) ?? mailboxId}`}
                        title="Unlink mailbox"
                        disabled={removingMailboxId === mailboxId}
                        onClick={() => removeMailbox(mailboxId)}
                        className="flex size-7 shrink-0 items-center justify-center rounded-md text-black/30 transition-colors hover:bg-red-500/10 hover:text-red-500 disabled:opacity-30 dark:text-white/30 dark:hover:text-red-500"
                      >
                        <Trash2 className="size-3.5" />
                      </button>
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="mt-3 text-[12px] leading-5 text-black/35 dark:text-white/30">
                  No mailboxes linked.
                </p>
              )}
            </div>

            {/* Test */}
            <div className="flex flex-wrap items-center gap-3">
              <button
                type="button"
                onClick={runTest}
                disabled={testing}
                className="inline-flex h-9 items-center gap-1.5 rounded-lg border border-black/[0.10] bg-white px-3 text-[13px] font-semibold text-black transition hover:bg-black/[0.03] active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-50 dark:border-white/[0.10] dark:bg-white/[0.04] dark:text-white dark:hover:bg-white/[0.07]"
              >
                {testing ? (
                  <Loader2 className="size-3.5 animate-spin" />
                ) : (
                  <Send className="size-3.5" />
                )}
                {testing ? "Sending…" : "Send test event"}
              </button>

              {testResult && (
                <span
                  className={`inline-flex items-center gap-1.5 rounded-lg px-2.5 py-1 font-mono text-[11px] font-medium ${
                    testResult.success
                      ? "bg-emerald-500/10 text-emerald-700 dark:text-emerald-300"
                      : "bg-red-500/10 text-red-600 dark:text-red-400"
                  }`}
                >
                  {testResult.success ? (
                    <Check className="size-3" />
                  ) : (
                    <AlertTriangle className="size-3" />
                  )}
                  HTTP {testResult.statusCode}
                </span>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function WebhookNameEditor({
  webhook,
  onUpdate,
}: {
  webhook: WebhookResponse
  onUpdate: (webhook: WebhookResponse) => void
}) {
  const [nameDraft, setNameDraft] = useState(webhook.name)
  const [saving, setSaving] = useState(false)

  const dirty = nameDraft.trim() !== webhook.name && nameDraft.trim() !== ""

  const save = async () => {
    if (!dirty || saving) return

    setSaving(true)
    try {
      const updated = await updateWebhook(webhook.id, { name: nameDraft.trim() })
      onUpdate(updated)
      toast.success("Webhook name updated")
    } catch (error) {
      toast.error("Could not update webhook name", {
        description: error instanceof Error ? error.message : "Something went wrong.",
      })
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="mt-2 flex items-center gap-2">
      <input
        type="text"
        value={nameDraft}
        onChange={(event) => setNameDraft(event.target.value)}
        onKeyDown={(event) => {
          if (event.key === "Enter") void save()
        }}
        placeholder="My webhook"
        className="h-9 min-w-0 flex-1 rounded-lg border border-black/[0.10] bg-[#fafaf9] px-3 text-[13px] outline-none transition placeholder:text-black/25 focus:border-black/25 focus:ring-4 focus:ring-black/[0.035] dark:border-white/[0.10] dark:bg-[#0b0c0c] dark:placeholder:text-white/20 dark:focus:border-white/25 dark:focus:ring-white/[0.035]"
      />
      <button
        type="button"
        onClick={() => void save()}
        disabled={!dirty || saving}
        className="inline-flex h-9 items-center rounded-lg border border-black/[0.10] bg-white px-3 text-[13px] font-semibold text-black transition hover:bg-black/[0.03] active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-40 dark:border-white/[0.10] dark:bg-white/[0.04] dark:text-white dark:hover:bg-white/[0.07]"
      >
        {saving ? "Saving…" : "Save"}
      </button>
    </div>
  )
}

function DeadLetters({
  deadLetters,
  onMarkSeen,
}: {
  deadLetters: WebhookDeadLetterResponse[]
  onMarkSeen: (id: string) => void
}) {
  const [open, setOpen] = useState(false)
  const [filter, setFilter] = useState<"unseen" | "all">("unseen")
  const [seenIds, setSeenIds] = useState<Set<string>>(new Set())
  const [markingId, setMarkingId] = useState<string | null>(null)

  const effectiveDeadLetters = deadLetters.filter(
    (deadLetter) => filter === "all" || !(deadLetter.seen || seenIds.has(deadLetter.id))
  )
  const unseenCount = deadLetters.filter(
    (deadLetter) => !(deadLetter.seen || seenIds.has(deadLetter.id))
  ).length

  const markSeen = async (id: string) => {
    if (markingId) return

    setMarkingId(id)
    try {
      await markWebhookDeadLetterSeen(id)
      setSeenIds((current) => new Set(current).add(id))
      onMarkSeen(id)
      toast.success("Marked as seen")
    } catch (error) {
      toast.error("Could not mark as seen", {
        description: error instanceof Error ? error.message : "Something went wrong.",
      })
    } finally {
      setMarkingId(null)
    }
  }

  return (
    <section className="mt-6 overflow-hidden rounded-2xl border border-black/[0.09] bg-white shadow-[0_12px_40px_rgba(0,0,0,0.05)] dark:border-white/[0.08] dark:bg-[#111313] dark:shadow-black/20">
      <button
        type="button"
        onClick={() => setOpen((current) => !current)}
        aria-expanded={open}
        className="flex w-full items-center justify-between gap-3 px-5 py-4 text-left transition-colors hover:bg-black/[0.02] dark:hover:bg-white/[0.03]"
      >
        <div className="flex items-center gap-3">
          <span className="flex size-8 items-center justify-center rounded-lg bg-red-500/[0.07] text-red-500">
            <AlertTriangle className="size-4" />
          </span>
          <div>
            <p className="text-sm font-semibold">Failed deliveries</p>
            <p className="mt-0.5 text-xs text-black/40 dark:text-white/35">
              {unseenCount} unread{" "}
              {unseenCount === 1 ? "delivery" : "deliveries"}{" "}
              {deadLetters.length > unseenCount
                ? `of ${deadLetters.length} total`
                : "that exhausted retries and were parked."}
            </p>
          </div>
        </div>

        <ChevronRight
          aria-hidden="true"
          className={`size-4 text-black/35 transition-transform duration-200 dark:text-white/30 ${
            open ? "rotate-90" : ""
          }`}
        />
      </button>

      {open && (
        <div className="animate-in fade-in border-t border-black/[0.07] dark:border-white/[0.07]">
          <div className="flex items-center justify-between gap-3 border-b border-black/[0.06] px-5 py-3 dark:border-white/[0.06]">
            <div className="flex items-center gap-1 rounded-lg bg-black/[0.04] p-0.5 dark:bg-white/[0.06]">
              {(["unseen", "all"] as const).map((option) => (
                <button
                  key={option}
                  type="button"
                  onClick={() => setFilter(option)}
                  className={`rounded-md px-3 py-1 text-[12px] font-medium transition ${
                    filter === option
                      ? "bg-white text-black shadow-sm dark:bg-[#1a1d1d] dark:text-white"
                      : "text-black/45 hover:text-black/70 dark:text-white/40 dark:hover:text-white/70"
                  }`}
                >
                  {option === "unseen" ? `Unread (${unseenCount})` : `All (${deadLetters.length})`}
                </button>
              ))}
            </div>

            <span className="text-[11px] text-black/35 dark:text-white/30">
              Newest first
            </span>
          </div>

          {effectiveDeadLetters.length === 0 ? (
            <div className="flex min-h-[120px] items-center justify-center px-6 text-center">
              <p className="text-[13px] text-black/35 dark:text-white/30">
                {filter === "unseen"
                  ? "No unread failed deliveries."
                  : "No failed deliveries."}
              </p>
            </div>
          ) : (
            <div className="divide-y divide-black/[0.05] dark:divide-white/[0.05]">
              {effectiveDeadLetters.map((deadLetter) => (
                <div key={deadLetter.id} className="px-5 py-3.5">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <div className="flex flex-wrap items-center gap-2">
                        {deadLetter.mailbox_address ? (
                          <span className="truncate font-mono text-[12px] font-semibold text-black/80 dark:text-white/80">
                            {deadLetter.mailbox_address}
                          </span>
                        ) : (
                          <span className="text-[12px] text-black/35 dark:text-white/30">
                            Deleted mailbox
                          </span>
                        )}

                        <span className="rounded-md bg-red-500/[0.08] px-1.5 py-0.5 font-mono text-[10px] font-medium text-red-700 dark:text-red-300">
                          {deadLetter.event}
                        </span>

                        {deadLetter.seen && (
                          <span className="rounded-md bg-black/[0.05] px-1.5 py-0.5 text-[10px] font-medium text-black/45 dark:bg-white/[0.07] dark:text-white/40">
                            Seen
                          </span>
                        )}
                      </div>

                      {deadLetter.message_subject || deadLetter.message_sender ? (
                        <p className="mt-1 truncate text-[12px] text-black/55 dark:text-white/50">
                          {deadLetter.message_subject ||
                            deadLetter.message_sender ||
                            "No subject"}{" "}
                          {deadLetter.message_sender &&
                            deadLetter.message_subject &&
                            "· "}
                          <span className="text-black/35 dark:text-white/30">
                            {deadLetter.message_sender}
                          </span>
                        </p>
                      ) : null}

                      <code className="mt-1 block truncate font-mono text-[11px] text-black/45 dark:text-white/35">
                        {deadLetter.url}
                      </code>
                    </div>

                    {!deadLetter.seen && (
                      <button
                        type="button"
                        onClick={() => void markSeen(deadLetter.id)}
                        disabled={markingId !== null}
                        className="inline-flex h-7 shrink-0 items-center gap-1 rounded-lg border border-black/[0.10] bg-white px-2.5 text-[12px] font-semibold text-black/70 transition hover:bg-black/[0.03] active:scale-[0.98] disabled:opacity-40 dark:border-white/[0.10] dark:bg-white/[0.04] dark:text-white/70 dark:hover:bg-white/[0.07]"
                      >
                        {markingId === deadLetter.id ? (
                          <Loader2 className="size-3 animate-spin" />
                        ) : (
                          <Check className="size-3" />
                        )}
                        Seen
                      </button>
                    )}
                  </div>

                  <p className="mt-1.5 font-mono text-[10px] leading-4 text-black/40 dark:text-white/35">
                    {deadLetter.reason} · {deadLetter.attempts} attempts ·{" "}
                    {deadLetter.created_at}
                  </p>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </section>
  )
}

function CreateWebhookDialog({
  mailboxes,
  onClose,
  onCreated,
}: {
  mailboxes: MailboxResponse[]
  onClose: () => void
  onCreated: (webhook: WebhookResponse) => void
}) {
  const [name, setName] = useState("")
  const [url, setUrl] = useState("")
  const [selectedEvents, setSelectedEvents] = useState<string[]>(["email.received"])
  const [selectedMailboxes, setSelectedMailboxes] = useState<Set<string>>(new Set())
  const [creating, setCreating] = useState(false)
  const [createdSecret, setCreatedSecret] = useState<string | null>(null)
  const [createdWebhook, setCreatedWebhook] = useState<WebhookResponse | null>(null)
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !createdSecret) {
        onClose()
      }
    }

    document.addEventListener("keydown", onKeyDown)
    return () => document.removeEventListener("keydown", onKeyDown)
  }, [createdSecret, onClose])

  const toggleEvent = (eventKey: string) => {
    setSelectedEvents((current) =>
      current.includes(eventKey)
        ? current.filter((key) => key !== eventKey)
        : [...current, eventKey]
    )
  }

  const toggleMailbox = (mailboxId: string) => {
    setSelectedMailboxes((current) => {
      const next = new Set(current)
      if (next.has(mailboxId)) {
        next.delete(mailboxId)
      } else {
        next.add(mailboxId)
      }
      return next
    })
  }

  const canSubmit =
    url.trim() !== "" &&
    isHttpUrl(url.trim()) &&
    selectedEvents.length > 0 &&
    selectedMailboxes.size > 0 &&
    !creating

  const submit = async () => {
    setCreating(true)

    try {
      const result = await createWebhook({
        name: name.trim(),
        url: url.trim(),
        events: selectedEvents,
        mailbox_ids: Array.from(selectedMailboxes),
      })

      setCreatedWebhook(result)
      setCreatedSecret(result.secret)
      toast.success("Webhook created")
    } catch (error) {
      toast.error("Could not create webhook", {
        description: error instanceof Error ? error.message : "Something went wrong.",
      })
      setCreating(false)
    }
  }

  const copySecret = async () => {
    if (!createdSecret) return

    try {
      await navigator.clipboard.writeText(createdSecret)
      setCopied(true)
      toast.success("Secret copied to clipboard")
      setTimeout(() => setCopied(false), 1800)
    } catch {
      toast.error("Could not copy secret")
    }
  }

  return (
    <DialogShell onBackdrop={() => (createdSecret ? undefined : onClose())}>
      {createdWebhook && createdSecret ? (
        <div className="relative w-full max-w-md rounded-2xl border border-black/10 bg-white p-6 shadow-2xl dark:border-white/10 dark:bg-[#111313]">
          <div className="flex items-start justify-between gap-3">
            <div>
              <h2 className="text-base font-semibold tracking-tight">
                Webhook created
              </h2>

              <p className="mt-1 text-[13px] leading-5 text-black/45 dark:text-white/40">
                Copy the signing secret now. Delivered webhooks include it in the{" "}
                <code className="rounded bg-black/[0.05] px-1 font-mono text-[11px] dark:bg-white/[0.07]">
                  X-Webhook-Signature
                </code>{" "}
                header so you can verify authenticity.
              </p>
            </div>

            <button
              type="button"
              aria-label="Close"
              onClick={onClose}
              className="flex size-8 shrink-0 items-center justify-center rounded-lg text-black/35 transition-colors hover:bg-black/[0.05] hover:text-black/70 dark:text-white/35 dark:hover:bg-white/[0.07] dark:hover:text-white/70"
            >
              <X className="size-4" />
            </button>
          </div>

          <div className="mt-5 rounded-xl border border-amber-500/20 bg-amber-500/[0.06] px-3.5 py-2.5 text-xs leading-5 text-amber-700 dark:text-amber-300">
            Save this secret securely — the server can&apos;t show it to you again.
          </div>

          <div className="mt-4 flex h-12 items-center overflow-hidden rounded-xl border border-black/[0.10] bg-[#fafaf9] dark:border-white/[0.10] dark:bg-[#0b0c0c]">
            <code className="min-w-0 flex-1 truncate px-3.5 font-mono text-[13px] font-medium text-black/75 dark:text-white/75">
              {createdSecret}
            </code>

            <button
              type="button"
              aria-label="Copy secret"
              title="Copy secret"
              onClick={copySecret}
              className="flex h-full shrink-0 items-center border-l border-black/[0.07] px-3 text-black/35 transition-colors hover:text-black/75 dark:border-white/[0.07] dark:text-white/35 dark:hover:text-white/75"
            >
              {copied ? (
                <Check className="size-4 text-emerald-600" />
              ) : (
                <Copy className="size-4" />
              )}
            </button>
          </div>

          <button
            type="button"
            onClick={() => {
              onCreated(createdWebhook)
              onClose()
            }}
            className="mt-5 h-10 w-full rounded-xl bg-black text-sm font-semibold text-white transition hover:bg-black/80 active:scale-[0.99] dark:bg-white dark:text-black dark:hover:bg-white/85"
          >
            Done
          </button>
        </div>
      ) : (
        <div className="relative w-full max-w-md rounded-2xl border border-black/10 bg-white p-6 shadow-2xl dark:border-white/10 dark:bg-[#111313]">
          <div className="flex items-start justify-between gap-3">
            <div>
              <h2 className="text-base font-semibold tracking-tight">
                Create webhook
              </h2>

              <p className="mt-1 text-[13px] leading-5 text-black/45 dark:text-white/40">
                Point us at an HTTPS endpoint to receive event callbacks.
              </p>
            </div>

            <button
              type="button"
              aria-label="Close"
              onClick={onClose}
              className="flex size-8 shrink-0 items-center justify-center rounded-lg text-black/35 transition-colors hover:bg-black/[0.05] hover:text-black/70 dark:text-white/35 dark:hover:bg-white/[0.07] dark:hover:text-white/70"
            >
              <X className="size-4" />
            </button>
          </div>

          <label
            htmlFor="webhook-name"
            className="mb-2 mt-5 block text-xs font-medium text-black/50 dark:text-white/40"
          >
            Name
          </label>
          <input
            id="webhook-name"
            type="text"
            autoComplete="off"
            spellCheck={false}
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder="Production receiver"
            className="h-11 w-full rounded-xl border border-black/[0.10] bg-[#fafaf9] px-3.5 text-[13px] outline-none transition placeholder:text-black/25 focus:border-black/25 focus:ring-4 focus:ring-black/[0.035] dark:border-white/[0.10] dark:bg-[#0b0c0c] dark:placeholder:text-white/20 dark:focus:border-white/25 dark:focus:ring-white/[0.035]"
          />

          <p className="mt-1.5 text-xs text-black/40 dark:text-white/35">
            Optional label. Falls back to the endpoint URL if empty.
          </p>

          <label
            htmlFor="webhook-url"
            className="mb-2 mt-5 block text-xs font-medium text-black/50 dark:text-white/40"
          >
            Endpoint URL
          </label>
          <input
            id="webhook-url"
            type="url"
            autoFocus
            autoComplete="off"
            spellCheck={false}
            value={url}
            onChange={(event) => setUrl(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Enter" && canSubmit) void submit()
            }}
            placeholder="https://example.com/webhooks/receiver"
            className="h-11 w-full rounded-xl border border-black/[0.10] bg-[#fafaf9] px-3.5 font-mono text-[13px] outline-none transition placeholder:text-black/25 focus:border-black/25 focus:ring-4 focus:ring-black/[0.035] dark:border-white/[0.10] dark:bg-[#0b0c0c] dark:placeholder:text-white/20 dark:focus:border-white/25 dark:focus:ring-white/[0.035]"
          />

          <p className="mt-1.5 text-xs text-black/40 dark:text-white/35">
            Must be a valid <code className="font-mono">https://</code> endpoint.
          </p>

          <p className="mb-2 mt-5 text-xs font-medium text-black/50 dark:text-white/40">
            Events
          </p>
          <div className="space-y-2">
            {EVENT_OPTIONS.map((event) => {
              const isOn = selectedEvents.includes(event.key)

              return (
                <button
                  key={event.key}
                  type="button"
                  role="checkbox"
                  aria-checked={isOn}
                  onClick={() => toggleEvent(event.key)}
                  className={`flex w-full items-center justify-between gap-3 rounded-xl border px-3 py-2.5 text-left transition ${
                    isOn
                      ? "border-black/[0.16] bg-black/[0.03] dark:border-white/[0.18] dark:bg-white/[0.05]"
                      : "border-black/[0.08] hover:border-black/[0.14] dark:border-white/[0.08] dark:hover:border-white/[0.15]"
                  }`}
                >
                  <span>
                    <code className="block font-mono text-[11px] font-medium text-black/70 dark:text-white/70">
                      {event.key}
                    </code>
                    <span className="mt-0.5 block text-[11px] text-black/40 dark:text-white/35">
                      {event.description}
                    </span>
                  </span>
                  <span
                    className={`flex size-4 shrink-0 items-center justify-center rounded-full transition ${
                      isOn
                        ? "bg-black text-white dark:bg-white dark:text-black"
                        : "border border-black/20 dark:border-white/20"
                    }`}
                  >
                    {isOn && <Check className="size-2.5" />}
                  </span>
                </button>
              )
            })}
          </div>

          <p className="mb-2 mt-5 text-xs font-medium text-black/50 dark:text-white/40">
            Mailboxes
          </p>

          {mailboxes.length === 0 ? (
            <div className="flex items-center justify-between gap-3 rounded-xl border border-black/[0.08] px-3.5 py-3 dark:border-white/[0.08]">
              <span className="flex items-center gap-2.5 text-[13px] text-black/40 dark:text-white/35">
                <Inbox className="size-4" />
                You need a mailbox to attach events to.
              </span>
              <Link
                href="/developer/mailbox"
                className="shrink-0 text-[13px] font-semibold text-black/70 underline-offset-4 hover:underline dark:text-white/70"
              >
                Create one
              </Link>
            </div>
          ) : (
            <div className="grid max-h-48 gap-1.5 overflow-y-auto pr-1">
              {mailboxes.map((mailbox) => {
                const isOn = selectedMailboxes.has(mailbox.id)

                return (
                  <button
                    key={mailbox.id}
                    type="button"
                    role="checkbox"
                    aria-checked={isOn}
                    onClick={() => toggleMailbox(mailbox.id)}
                    className={`flex items-center justify-between gap-3 rounded-lg border px-3 py-2 text-left transition ${
                      isOn
                        ? "border-black/[0.16] bg-black/[0.03] dark:border-white/[0.18] dark:bg-white/[0.05]"
                        : "border-black/[0.08] hover:border-black/[0.14] dark:border-white/[0.08] dark:hover:border-white/[0.15]"
                    }`}
                  >
                    <span className="truncate font-mono text-[12px] text-black/70 dark:text-white/70">
                      {mailbox.address}
                    </span>
                    <span
                      className={`flex size-4 shrink-0 items-center justify-center rounded-full transition ${
                        isOn
                          ? "bg-black text-white dark:bg-white dark:text-black"
                          : "border border-black/20 dark:border-white/20"
                      }`}
                    >
                      {isOn && <Check className="size-2.5" />}
                    </span>
                  </button>
                )
              })}
            </div>
          )}

          <div className="mt-5 flex items-center justify-end gap-2">
            <button
              type="button"
              onClick={onClose}
              className="h-10 rounded-xl px-4 text-sm font-medium text-black/55 transition hover:bg-black/[0.04] dark:text-white/50 dark:hover:bg-white/[0.06]"
            >
              Cancel
            </button>

            <button
              type="button"
              onClick={() => {
                void submit()
              }}
              disabled={!canSubmit}
              className="h-10 rounded-xl bg-black px-4 text-sm font-semibold text-white transition hover:bg-black/80 active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-30 dark:bg-white dark:text-black dark:hover:bg-white/85"
            >
              {creating ? "Creating…" : "Create webhook"}
            </button>
          </div>
        </div>
      )}
    </DialogShell>
  )
}

function DeleteWebhookDialog({
  target,
  onCancel,
  onDeleted,
}: {
  target: WebhookResponse
  onCancel: () => void
  onDeleted: () => void
}) {
  const [confirmText, setConfirmText] = useState("")
  const [deleting, setDeleting] = useState(false)

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !deleting) {
        onCancel()
      }
    }

    document.addEventListener("keydown", onKeyDown)
    return () => document.removeEventListener("keydown", onKeyDown)
  }, [deleting, onCancel])

  const matches = confirmText === target.name
  const canDelete = matches && !deleting

  const deleteIt = async () => {
    setDeleting(true)

    try {
      await deleteWebhook(target.id)
      toast.success("Webhook deleted")
      onDeleted()
    } catch (error) {
      toast.error("Could not delete webhook", {
        description: error instanceof Error ? error.message : "Something went wrong.",
      })
      setDeleting(false)
    }
  }

  return (
    <DialogShell onBackdrop={() => (deleting ? undefined : onCancel())}>
      <div className="relative w-full max-w-md rounded-2xl border border-black/10 bg-white p-6 shadow-2xl dark:border-white/10 dark:bg-[#111313]">
        <h2 className="flex items-center gap-2 text-base font-semibold tracking-tight">
          <AlertTriangle className="size-4 text-red-500" />
          Delete webhook?
        </h2>

        <p className="mt-1.5 text-[13px] leading-5 text-black/45 dark:text-white/40">
          This action is permanent. The endpoint will stop receiving events and
          its signing secret can&apos;t be recovered.
        </p>

        <div className="mt-4 rounded-xl border border-red-500/15 bg-red-500/[0.05] px-3.5 py-3">
          <p className="truncate text-[13px] font-semibold text-red-700/90 dark:text-red-300/90">
            {target.name}
          </p>
          <p className="mt-0.5 truncate font-mono text-[11px] text-red-700/60 dark:text-red-300/50">
            {target.url}
          </p>
        </div>

        <label
          htmlFor="webhook-delete-confirm"
          className="mb-2 mt-5 block text-xs font-medium text-black/50 dark:text-white/40"
        >
          Type <code className="font-mono">{target.name || "the webhook name"}</code> to confirm
        </label>
        <input
          id="webhook-delete-confirm"
          type="text"
          autoFocus
          autoComplete="off"
          spellCheck={false}
          value={confirmText}
          onChange={(event) => setConfirmText(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === "Enter" && canDelete) void deleteIt()
          }}
          placeholder={`Type "${target.name}"`}
          className="h-11 w-full rounded-xl border border-black/[0.10] bg-[#fafaf9] px-3.5 font-mono text-[13px] outline-none transition placeholder:text-black/25 focus:border-red-500/40 focus:ring-4 focus:ring-red-500/[0.06] dark:border-white/[0.10] dark:bg-[#0b0c0c] dark:placeholder:text-white/20 dark:focus:border-red-500/40 dark:focus:ring-red-500/[0.06]"
        />

        <div className="mt-5 flex items-center justify-end gap-2">
          <button
            type="button"
            onClick={onCancel}
            className="h-10 rounded-xl px-4 text-sm font-medium text-black/55 transition hover:bg-black/[0.04] dark:text-white/50 dark:hover:bg-white/[0.06]"
          >
            Cancel
          </button>

          <button
            type="button"
            onClick={() => {
              void deleteIt()
            }}
            disabled={!canDelete}
            className="h-10 rounded-xl bg-red-600 px-4 text-sm font-semibold text-white transition hover:bg-red-600/85 active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-30"
          >
            {deleting ? "Deleting…" : "Yes, delete webhook"}
          </button>
        </div>
      </div>
    </DialogShell>
  )
}

function DialogShell({
  children,
  onBackdrop,
}: {
  children: React.ReactNode
  onBackdrop: () => void
}) {
  return (
    <div
      role="dialog"
      aria-modal="true"
      className="fixed inset-0 z-50 flex items-center justify-center p-4"
    >
      <div
        aria-hidden="true"
        className="absolute inset-0 bg-black/40 backdrop-blur-sm"
        onClick={onBackdrop}
      />

      <div className="relative w-full max-w-md">{children}</div>
    </div>
  )
}

function isHttpUrl(value: string): boolean {
  try {
    const parsed = new URL(value)
    return parsed.protocol === "https:" || parsed.protocol === "http:"
  } catch {
    return false
  }
}