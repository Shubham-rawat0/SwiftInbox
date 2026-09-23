"use client"

import {
  createDeveloperApiKey,
  getDeveloper,
  isApiError,
  listDeveloperApiKeys,
  revokeDeveloperApiKey,
} from "@/lib/api"
import type {
  ApiKeyUsageResponse,
  CreateApiKeyResponse,
  DeveloperDetailsResponse,
  NullableTime,
} from "@/lib/types"
import {
  Check,
  ChevronRight,
  Copy,
  KeyRound,
  Plus,
  Trash2,
  X,
} from "lucide-react"
import Link from "next/link"
import { useEffect, useState } from "react"
import { toast } from "sonner"

const CURRENT_MONTH = new Date().toLocaleDateString("en-US", {
  month: "long",
  year: "numeric",
})

// Shared grid so the header row and every key row line up.
const ROW_GRID =
  "grid-cols-[16px_minmax(0,1fr)] sm:grid-cols-[16px_minmax(0,1.1fr)_minmax(0,1.2fr)_96px_96px]"

type UsagePoint = { label: string; value: number }

const DOT_CLASSES = [
  "bg-black dark:bg-white",
  "bg-black/60 dark:bg-white/60",
  "bg-black/30 dark:bg-white/30",
]

export default function ApiKeysPage() {
  return <ApiKeysManager />
}

function ApiKeysManager() {
  const [keys, setKeys] = useState<ApiKeyUsageResponse[] | null>(null)
  const [profile, setProfile] = useState<DeveloperDetailsResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [signedOut, setSignedOut] = useState(false)
  const [failed, setFailed] = useState(false)
  const [reloadKey, setReloadKey] = useState(0)
  const [showCreate, setShowCreate] = useState(false)

  useEffect(() => {
    let isActive = true

    async function loadKeys() {
      try {
        const [keysResult, profileResult] = await Promise.all([
          listDeveloperApiKeys(),
          getDeveloper(),
        ])

        if (!isActive) return

        setKeys(keysResult.filter((key) => !isRevoked(key)))
        setProfile(profileResult)
        setFailed(false)
      } catch (error) {
        if (!isActive) return

        setKeys([])

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

    void loadKeys()

    return () => {
      isActive = false
    }
  }, [reloadKey])

  const onCreated = () => {
    setShowCreate(false)
    setReloadKey((value) => value + 1)
  }

  const onRevoked = (id: string) => {
    setKeys((current) =>
      current === null ? current : current.filter((key) => key.ID !== id)
    )
  }

  const showContent = !loading && !signedOut && !failed

  return (
    <div className="min-h-[calc(100svh-4rem)] bg-[#f8f8f6] text-[#171717] dark:bg-[#0b0c0c] dark:text-white">
      <main className="mx-auto w-full max-w-5xl px-5 py-6 sm:px-8 sm:py-8">
        <div className="animate-in fade-in duration-300 ease-out">
          <div className="mb-6 flex flex-wrap items-start justify-between gap-4">
            <div className="max-w-xl">
              <h1 className="font-heading text-2xl font-semibold tracking-[-0.035em] sm:text-3xl">
                API keys
              </h1>

              <p className="mt-1.5 text-[14px] leading-5 text-black/45 dark:text-white/40">
                Create and manage secret keys used to authenticate requests to
                the Mailbox API. Keys inherit the quotas of your developer
                account. Select a key to see its usage for {CURRENT_MONTH}.
              </p>
            </div>

            {showContent && (
              <button
                type="button"
                onClick={() => setShowCreate(true)}
                className="inline-flex h-10 shrink-0 items-center gap-1.5 rounded-xl bg-black px-4 text-[13px] font-semibold text-white transition hover:bg-black/80 active:scale-[0.99] dark:bg-white dark:text-black dark:hover:bg-white/85"
              >
                <Plus className="size-4" />
                Create new secret key
              </button>
            )}
          </div>

          {loading ? (
            <div className="flex min-h-[320px] items-center justify-center">
              <div className="flex items-center gap-2.5 text-sm text-muted-foreground">
                <div className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                Loading API keys
              </div>
            </div>
          ) : signedOut ? (
            <div className="flex min-h-[320px] items-center justify-center px-6">
              <div className="text-center">
                <p className="text-sm font-medium">
                  Sign in to view your API keys
                </p>

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
                <p className="text-sm font-medium">
                  Couldn&apos;t load your API keys
                </p>

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
              {profile && <AccountUsageSection profile={profile} />}

              <KeysTable
                keys={keys ?? []}
                onCreate={() => setShowCreate(true)}
                onRevoked={onRevoked}
              />
            </>
          )}
        </div>
      </main>

      {showCreate && (
        <CreateKeyDialog
          onClose={() => setShowCreate(false)}
          onCreated={onCreated}
        />
      )}
    </div>
  )
}

function AccountUsageSection({
  profile,
}: {
  profile: DeveloperDetailsResponse
}) {
  const meters = [
    {
      label: "API requests",
      used: profile.ApiRequests,
      limit: profile.ApiQuota,
    },
    {
      label: "Mailboxes",
      used: profile.MailboxRequests,
      limit: profile.MailboxQuota,
    },
    {
      label: "Messages",
      used: profile.MessagesRequests,
      limit: profile.MessageQuota,
    },
  ]

  return (
    <section className="mb-6 overflow-hidden rounded-2xl border border-black/[0.09] bg-white shadow-[0_12px_40px_rgba(0,0,0,0.05)] dark:border-white/[0.08] dark:bg-[#111313] dark:shadow-black/20">
      <div className="flex items-center justify-between px-5 py-4">
        <div>
          <p className="text-sm font-semibold">Account usage</p>

          <p className="mt-0.5 text-xs text-black/40 dark:text-white/35">
            Total across all keys for {CURRENT_MONTH}.
          </p>
        </div>

        <span className="rounded-md bg-black/[0.04] px-2 py-1 font-mono text-[10px] font-medium tracking-wide text-black/40 dark:bg-white/[0.05] dark:text-white/35">
          THIS MONTH
        </span>
      </div>

      <div className="border-t border-black/[0.07] px-5 py-5 dark:border-white/[0.07]">
        <div className="grid gap-x-10 gap-y-5 sm:grid-cols-3">
          {meters.map((meter) => (
            <UsageMeter key={meter.label} {...meter} />
          ))}
        </div>
      </div>

      <div className="border-t border-black/[0.07] bg-black/[0.015] px-5 py-3 dark:border-white/[0.07] dark:bg-white/[0.015]">
        <p className="text-xs leading-5 text-black/40 dark:text-white/35">
          Quotas apply to your whole account and reset at the start of each
          month.
        </p>
      </div>
    </section>
  )
}

function UsageMeter({
  label,
  used,
  limit,
}: {
  label: string
  used: number
  limit: number
}) {
  const over = limit > 0 && used > limit
  const percentage =
    limit > 0 ? Math.min(100, Math.round((used / limit) * 100)) : 0
  const fillWidth = used > 0 ? `${Math.max(percentage, 2)}%` : "0%"

  return (
    <div>
      <div className="flex items-baseline justify-between gap-2">
        <span className="text-[13px] font-medium">{label}</span>

        <span
          className={`shrink-0 tabular-nums text-[13px] ${
            over
              ? "font-semibold text-red-600 dark:text-red-400"
              : "text-black/55 dark:text-white/50"
          }`}
        >
          {used.toLocaleString()}
          <span className="ml-1 text-black/35 dark:text-white/30">
            / {limit.toLocaleString()}
          </span>
        </span>
      </div>

      <div className="mt-2.5 h-1.5 overflow-hidden rounded-full bg-black/[0.07] dark:bg-white/[0.08]">
        <div
          role="progressbar"
          aria-valuenow={percentage}
          aria-valuemin={0}
          aria-valuemax={100}
          className={`h-full rounded-full transition-all ${
            over
              ? "bg-red-500"
              : "bg-black dark:bg-white"
          }`}
          style={{ width: fillWidth }}
        />
      </div>
    </div>
  )
}

function KeysTable({
  keys,
  onCreate,
  onRevoked,
}: {
  keys: ApiKeyUsageResponse[]
  onCreate: () => void
  onRevoked: (id: string) => void
}) {
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [revokeTarget, setRevokeTarget] =
    useState<ApiKeyUsageResponse | null>(null)

  return (
    <section className="overflow-hidden rounded-2xl border border-black/[0.09] bg-white shadow-[0_12px_40px_rgba(0,0,0,0.05)] dark:border-white/[0.08] dark:bg-[#111313] dark:shadow-black/20">
      {keys.length === 0 ? (
        <div className="flex min-h-[260px] flex-col items-center justify-center gap-3 px-6 text-center">
          <span className="flex size-12 items-center justify-center rounded-full bg-black/[0.04] dark:bg-white/[0.05]">
            <KeyRound className="size-5 text-black/35 dark:text-white/30" />
          </span>

          <div>
            <p className="text-sm font-medium">No secret keys yet</p>
            <p className="mt-1 text-[13px] text-black/40 dark:text-white/35">
              Create a key to call the API programmatically and track its
              usage.
            </p>
          </div>

          <button
            type="button"
            onClick={onCreate}
            className="mt-2 inline-flex h-9 items-center gap-1.5 rounded-lg bg-black px-3 text-[13px] font-semibold text-white transition hover:bg-black/80 active:scale-[0.99] dark:bg-white dark:text-black dark:hover:bg-white/85"
          >
            <Plus className="size-4" />
            Create new secret key
          </button>
        </div>
      ) : (
        <>
          {/* Column headers */}
          <div className="hidden border-b border-black/[0.07] bg-black/[0.02] text-xs font-medium text-black/45 dark:border-white/[0.07] dark:bg-white/[0.02] dark:text-white/40 sm:flex">
            <div className={`grid flex-1 gap-4 px-5 py-2.5 ${ROW_GRID}`}>
              <span aria-hidden="true" />
              <span>Name</span>
              <span>Secret key</span>
              <span>Created</span>
              <span>Last used</span>
            </div>
            <div className="w-12 shrink-0" aria-hidden="true" />
          </div>

          <div className="divide-y divide-black/[0.06] dark:divide-white/[0.06]">
            {keys.map((key) => (
              <KeyRow
                key={key.ID}
                apiKey={key}
                expanded={expandedId === key.ID}
                onToggle={() =>
                  setExpandedId((current) =>
                    current === key.ID ? null : key.ID
                  )
                }
                onRevoke={() => setRevokeTarget(key)}
              />
            ))}
          </div>
        </>
      )}

      {revokeTarget && (
        <RevokeKeyDialog
          key={revokeTarget.ID}
          target={revokeTarget}
          onCancel={() => setRevokeTarget(null)}
          onRevoked={() => {
            onRevoked(revokeTarget.ID)
            setExpandedId((current) =>
              current === revokeTarget.ID ? null : current
            )
            setRevokeTarget(null)
          }}
        />
      )}
    </section>
  )
}

function KeyRow({
  apiKey,
  expanded,
  onToggle,
  onRevoke,
}: {
  apiKey: ApiKeyUsageResponse
  expanded: boolean
  onToggle: () => void
  onRevoke: () => void
}) {
  const revoked = isRevoked(apiKey)
  const panelId = `usage-${apiKey.ID}`
  const lastUsed = formatNullableTime(apiKey.LastUsedAt)

  return (
    <div className={expanded ? "bg-black/[0.015] dark:bg-white/[0.015]" : ""}>
      <div className="flex items-center">
        <button
          type="button"
          onClick={onToggle}
          aria-expanded={expanded}
          aria-controls={panelId}
          className={`grid min-w-0 flex-1 items-center gap-4 px-5 py-4 text-left transition-colors hover:bg-black/[0.02] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-black/20 dark:hover:bg-white/[0.03] dark:focus-visible:ring-white/25 ${ROW_GRID}`}
        >
          <ChevronRight
            aria-hidden="true"
            className={`size-4 text-black/35 transition-transform duration-200 dark:text-white/30 ${
              expanded ? "rotate-90" : ""
            }`}
          />

          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <p
                className={`truncate text-sm font-medium ${
                  revoked ? "text-black/40 dark:text-white/40" : ""
                }`}
              >
                {apiKey.Name}
              </p>

              {revoked && (
                <span className="shrink-0 rounded-md bg-red-500/10 px-1.5 py-0.5 text-[11px] font-medium text-red-600 dark:text-red-400">
                  Revoked
                </span>
              )}
            </div>

            {/* Secret key shows under the name on small screens */}
            <p className="mt-1 truncate font-mono text-[11px] text-black/35 dark:text-white/30 sm:hidden">
              {maskedKey(apiKey.ID)}
            </p>
          </div>

          <p className="hidden truncate font-mono text-[12px] text-black/45 dark:text-white/40 sm:block">
            {maskedKey(apiKey.ID)}
          </p>

          <p className="hidden text-[13px] text-black/55 dark:text-white/50 sm:block">
            {formatNullableTime(apiKey.CreatedAt)}
          </p>

          <p className="hidden text-[13px] text-black/55 dark:text-white/50 sm:block">
            {lastUsed === "Never" ? "Never" : lastUsed}
          </p>
        </button>

        <div className="flex w-12 shrink-0 justify-center">
          <button
            type="button"
            aria-label={`Revoke ${apiKey.Name}`}
            title="Revoke API key"
            disabled={revoked}
            onClick={onRevoke}
            className="flex size-9 items-center justify-center rounded-lg text-black/30 transition-colors hover:bg-red-500/10 hover:text-red-500 disabled:cursor-not-allowed disabled:opacity-30 dark:text-white/30 dark:hover:text-red-500"
          >
            <Trash2 className="size-4" />
          </button>
        </div>
      </div>

      {expanded && <KeyUsagePanel id={panelId} apiKey={apiKey} />}
    </div>
  )
}

function KeyUsagePanel({
  id,
  apiKey,
}: {
  id: string
  apiKey: ApiKeyUsageResponse
}) {
  const points: UsagePoint[] = [
    { label: "API", value: apiKey.ApiRequests },
    { label: "Mailbox", value: apiKey.MailboxRequests },
    { label: "Messages", value: apiKey.MessageRequests },
  ]
  const total = points.reduce((sum, point) => sum + point.value, 0)

  return (
    <div
      id={id}
      className="animate-in fade-in slide-in-from-top-1 border-t border-black/[0.06] px-5 pb-6 pt-5 duration-200 dark:border-white/[0.06] sm:pl-[52px]"
    >
      <div className="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
        <p className="text-sm font-semibold">Usage</p>
        <p className="text-xs text-black/40 dark:text-white/35">
          Requests recorded for {CURRENT_MONTH}
        </p>
      </div>

      <div className="mt-4 grid gap-6 sm:grid-cols-[minmax(0,220px)_minmax(0,1fr)] sm:items-center">
        <div>
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-semibold tracking-tight">
              {total.toLocaleString()}
            </span>
            <span className="text-xs text-black/40 dark:text-white/35">
              requests this month
            </span>
          </div>

          <ul className="mt-4 space-y-2.5">
            {points.map((point, index) => {
              const share = total === 0 ? 0 : (point.value / total) * 100

              return (
                <li
                  key={point.label}
                  className="flex items-center justify-between gap-3 text-xs"
                >
                  <span className="flex items-center gap-2 text-black/55 dark:text-white/50">
                    <span
                      aria-hidden="true"
                      className={`size-2 rounded-full ${DOT_CLASSES[index]}`}
                    />
                    {point.label}
                  </span>

                  <span className="tabular-nums text-black/70 dark:text-white/70">
                    {point.value.toLocaleString()}
                    <span className="ml-1.5 text-black/35 dark:text-white/30">
                      {share.toFixed(share % 1 === 0 ? 0 : 1)}%
                    </span>
                  </span>
                </li>
              )
            })}
          </ul>
        </div>

        <div className="max-w-md">
          <UsageChart points={points} />
        </div>
      </div>
    </div>
  )
}

function UsageChart({ points }: { points: UsagePoint[] }) {
  const W = 320
  const H = 176
  const PAD_TOP = 28
  const PAD_BOTTOM = 30
  const plotH = H - PAD_TOP - PAD_BOTTOM
  const max = Math.max(1, ...points.map((point) => point.value))
  const slot = W / points.length
  const barWidth = Math.min(52, slot * 0.52)

  return (
    <svg
      viewBox={`0 0 ${W} ${H}`}
      className="h-auto w-full"
      role="img"
      aria-label="API key usage by category"
    >
      {[0.25, 0.5, 0.75, 1].map((fraction) => (
        <line
          key={fraction}
          x1={0}
          x2={W}
          y1={PAD_TOP + plotH * (1 - fraction)}
          y2={PAD_TOP + plotH * (1 - fraction)}
          className="stroke-black/[0.07] dark:stroke-white/[0.07]"
          strokeWidth={1}
        />
      ))}

      {points.map((point, index) => {
        const height = Math.round((point.value / max) * plotH)
        const y = PAD_TOP + plotH - height
        const x = slot * index + (slot - barWidth) / 2
        const opacity =
          index === 0 ? "opacity-100" : index === 1 ? "opacity-60" : "opacity-30"

        return (
          <g key={point.label}>
            <rect
              x={x}
              y={y}
              width={barWidth}
              height={height}
              rx={6}
              className={`fill-black dark:fill-white ${opacity}`}
            >
              <title>{`${point.label}: ${point.value.toLocaleString()} requests`}</title>
            </rect>

            <text
              x={slot * index + slot / 2}
              y={y - 8}
              textAnchor="middle"
              className="fill-black/70 text-[11px] font-medium dark:fill-white/70"
            >
              {point.value.toLocaleString()}
            </text>

            <text
              x={slot * index + slot / 2}
              y={H - 11}
              textAnchor="middle"
              className="fill-black/40 text-[11px] dark:fill-white/35"
            >
              {point.label}
            </text>
          </g>
        )
      })}
    </svg>
  )
}

function CreateKeyDialog({
  onClose,
  onCreated,
}: {
  onClose: () => void
  onCreated: () => void
}) {
  const [name, setName] = useState("")
  const [creating, setCreating] = useState(false)
  const [created, setCreated] = useState<CreateApiKeyResponse | null>(null)
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !created) {
        onClose()
      }
    }

    document.addEventListener("keydown", onKeyDown)
    return () => document.removeEventListener("keydown", onKeyDown)
  }, [created, onClose])

  const submit = async () => {
    setCreating(true)

    try {
      const result = await createDeveloperApiKey(name.trim() || "default")
      setCreated(result)
      toast.success("Secret key created")
    } catch (error) {
      toast.error("Could not create API key", {
        description:
          error instanceof Error ? error.message : "Something went wrong.",
      })
    } finally {
      setCreating(false)
    }
  }

  const copyKey = async () => {
    if (!created) return

    try {
      await navigator.clipboard.writeText(created.api_key)
      setCopied(true)
      toast.success("API key copied to clipboard")
      setTimeout(() => setCopied(false), 1800)
    } catch {
      toast.error("Could not copy API key")
    }
  }

  return (
    <DialogShell onBackdrop={() => (created ? undefined : onClose())}>
      {created ? (
        <div className="relative w-full max-w-md rounded-2xl border border-black/10 bg-white p-6 shadow-2xl dark:border-white/10 dark:bg-[#111313]">
          <div className="flex items-start justify-between gap-3">
            <div>
              <h2 className="text-base font-semibold tracking-tight">
                Secret key created
              </h2>

              <p className="mt-1 text-[13px] leading-5 text-black/45 dark:text-white/40">
                Copy this key now. For security, it won&apos;t be shown again.
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
            Save this key securely — you won&apos;t be able to retrieve it
            again.
          </div>

          <div className="mt-4 flex h-12 items-center overflow-hidden rounded-xl border border-black/[0.10] bg-[#fafaf9] dark:border-white/[0.10] dark:bg-[#0b0c0c]">
            <code className="min-w-0 flex-1 truncate px-3.5 font-mono text-[13px] font-medium text-black/75 dark:text-white/75">
              {created.api_key}
            </code>

            <button
              type="button"
              aria-label="Copy API key"
              title="Copy API key"
              onClick={copyKey}
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
            onClick={onCreated}
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
                Create new secret key
              </h2>

              <p className="mt-1 text-[13px] leading-5 text-black/45 dark:text-white/40">
                Name the key so you can recognize it later.
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
            htmlFor="api-key-name"
            className="mb-2 mt-5 block text-xs font-medium text-black/50 dark:text-white/40"
          >
            Key name
          </label>

          <input
            id="api-key-name"
            type="text"
            autoFocus
            autoComplete="off"
            value={name}
            onChange={(event) => setName(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Enter") void submit()
            }}
            placeholder="default"
            className="h-11 w-full rounded-xl border border-black/[0.10] bg-[#fafaf9] px-3.5 text-sm font-medium outline-none transition placeholder:text-black/25 focus:border-black/25 focus:ring-4 focus:ring-black/[0.035] dark:border-white/[0.10] dark:bg-[#0b0c0c] dark:placeholder:text-white/20 dark:focus:border-white/25 dark:focus:ring-white/[0.035]"
          />

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
              disabled={creating}
              className="h-10 rounded-xl bg-black px-4 text-sm font-semibold text-white transition hover:bg-black/80 active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-30 dark:bg-white dark:text-black dark:hover:bg-white/85"
            >
              {creating ? "Creating..." : "Create secret key"}
            </button>
          </div>
        </div>
      )}
    </DialogShell>
  )
}

function RevokeKeyDialog({
  target,
  onCancel,
  onRevoked,
}: {
  target: ApiKeyUsageResponse
  onCancel: () => void
  onRevoked: () => void
}) {
  const [confirmText, setConfirmText] = useState("")
  const [revoking, setRevoking] = useState(false)

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !revoking) {
        onCancel()
      }
    }

    document.addEventListener("keydown", onKeyDown)
    return () => document.removeEventListener("keydown", onKeyDown)
  }, [revoking, onCancel])

  const matches = confirmText === target.Name
  const canRevoke = matches && !revoking

  const revoke = async () => {
    setRevoking(true)

    try {
      await revokeDeveloperApiKey(target.ID)
      toast.success(`Revoked “${target.Name}”`)
      onRevoked()
    } catch (error) {
      toast.error("Could not revoke API key", {
        description:
          error instanceof Error ? error.message : "Something went wrong.",
      })
      setRevoking(false)
    }
  }

  return (
    <DialogShell onBackdrop={() => (revoking ? undefined : onCancel())}>
      <div className="relative w-full max-w-md rounded-2xl border border-black/10 bg-white p-6 shadow-2xl dark:border-white/10 dark:bg-[#111313]">
        <h2 className="text-base font-semibold tracking-tight">
          Revoke API key?
        </h2>

        <p className="mt-1.5 text-[13px] leading-5 text-black/45 dark:text-white/40">
          This will permanently revoke{" "}
          <span className="font-medium text-black/70 dark:text-white/70">
            {target.Name}
          </span>{" "}
          ({maskedKey(target.ID)}). Any requests using this key will
          immediately stop working.
        </p>

        <div className="mt-4 rounded-xl border border-red-500/15 bg-red-500/[0.05] px-3.5 py-3">
          <p className="truncate font-mono text-[13px] font-semibold text-red-700/90 dark:text-red-300/90">
            {target.Name}
          </p>
          <p className="mt-0.5 truncate font-mono text-[11px] text-red-700/60 dark:text-red-300/50">
            {maskedKey(target.ID)}
          </p>
        </div>

        <label
          htmlFor="api-key-revoke-confirm"
          className="mb-2 mt-5 block text-xs font-medium text-black/50 dark:text-white/40"
        >
          Type <code className="font-mono">{target.Name}</code> to confirm
        </label>
        <input
          id="api-key-revoke-confirm"
          type="text"
          autoFocus
          autoComplete="off"
          spellCheck={false}
          value={confirmText}
          onChange={(event) => setConfirmText(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === "Enter" && canRevoke) void revoke()
          }}
          placeholder={`Type "${target.Name}"`}
          className="h-11 w-full rounded-xl border border-black/[0.10] bg-[#fafaf9] px-3.5 text-sm outline-none transition placeholder:text-black/25 focus:border-red-500/40 focus:ring-4 focus:ring-red-500/[0.06] dark:border-white/[0.10] dark:bg-[#0b0c0c] dark:placeholder:text-white/20 dark:focus:border-red-500/40 dark:focus:ring-red-500/[0.06]"
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
              void revoke()
            }}
            disabled={!canRevoke}
            className="h-10 rounded-xl bg-red-600 px-4 text-sm font-semibold text-white transition hover:bg-red-600/85 active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-30"
          >
            {revoking ? "Revoking..." : "Yes, revoke"}
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

function maskedKey(id: string): string {
  const compact = id.replace(/-/g, "")
  const prefix = compact.slice(0, 6)
  const suffix = compact.slice(-6)
  return `dev_${prefix}…${suffix}`
}

function isRevoked(key: ApiKeyUsageResponse): boolean {
  return Boolean(key.RevokedAt?.Valid)
}

function formatNullableTime(value: NullableTime | null | undefined): string {
  if (!value?.Valid || !value.Time) return "Never"

  const time = new Date(value.Time).getTime()
  if (Number.isNaN(time) || time <= 0) return "recently"

  return new Date(value.Time).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  })
}