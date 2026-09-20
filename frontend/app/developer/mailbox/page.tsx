"use client"

import { MailboxNavigation } from "@/components/developer/MailboxNavigation"
import { DeveloperSessionGate } from "@/components/developer/DeveloperSessionGate"
import { deleteDeveloperMailbox, isApiError, listDeveloperMailboxes } from "@/lib/api"
import { MailboxResponse } from "@/lib/types"
import { Check, Copy, Inbox, Trash2 } from "lucide-react"
import Link from "next/link"
import { useEffect, useState } from "react"
import { toast } from "sonner"

const MAIL_DOMAIN =
  process.env.NEXT_PUBLIC_MAIL_DOMAIN || "temp.mail.at"

export default function DeveloperPage() {
  return (
    <DeveloperSessionGate>
      <DeveloperMailboxList />
    </DeveloperSessionGate>
  )
}

function DeveloperMailboxList() {
  const [mailboxes, setMailboxes] = useState<MailboxResponse[] | null>(null)
  const [loading, setLoading] = useState(true)
  const [signedOut, setSignedOut] = useState(false)
  const [failed, setFailed] = useState(false)
  const [reloadKey, setReloadKey] = useState(0)
  const [copiedAddress, setCopiedAddress] = useState<string | null>(null)

  useEffect(() => {
    let isActive = true

    async function loadMailboxes() {
      try {
        const result = await listDeveloperMailboxes()

        if (!isActive) return

        setMailboxes(result)
        setFailed(false)
      } catch (error) {
        if (!isActive) return

        setMailboxes([])

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

    void loadMailboxes()

    return () => {
      isActive = false
    }
  }, [reloadKey])

  const copyMailbox = async (address: string) => {
    try {
      await navigator.clipboard.writeText(address)

      setCopiedAddress(address)
      toast.success("Email copied to clipboard!")
      setTimeout(() => setCopiedAddress(null), 1800)
    } catch {
      toast.error("Failed to copy email")
    }
  }

  const removeMailbox = async (address: string) => {
    try {
      await deleteDeveloperMailbox(address)

      setMailboxes((current) =>
        current === null
          ? current
          : current.filter((mailbox) => mailbox.address !== address)
      )
    } catch {
      toast.error("Could not delete mailbox")
    }
  }

  if (loading) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <div className="flex items-center gap-2.5 text-sm text-muted-foreground">
          <div className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
          Loading mailboxes
        </div>
      </div>
    )
  }

  if (signedOut) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center px-6">
        <div className="text-center">
          <p className="text-sm font-medium">
            Sign in to view your mailboxes
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
    )
  }

  if (failed) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center px-6">
        <div className="text-center">
          <p className="text-sm font-medium">
            Couldn&apos;t load your mailboxes
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
    )
  }

  return (
    <div className="min-h-[calc(100svh-4rem)] bg-[#f8f8f6] text-[#171717] dark:bg-[#0b0c0c] dark:text-white">
      <main className="mx-auto w-full max-w-4xl px-5 py-6 sm:px-8 sm:py-8">
        <div className="animate-in fade-in duration-300 ease-out">
          <div className="mb-6">
            <h1 className="font-heading text-2xl font-semibold tracking-[-0.035em] sm:text-3xl">
              User Mailboxes
            </h1>

            <p className="mt-1.5 text-[14px] leading-5 text-black/45 dark:text-white/40">
              View and manage mailboxes created for your account.
            </p>
          </div>

          <MailboxNavigation />

          <div className="mt-6">
            {mailboxes === null ? (
              <div className="flex min-h-[320px] items-center justify-center">
                <p className="text-[13px] text-black/35 dark:text-white/30">
                  Loading mailboxes…
                </p>
              </div>
            ) : mailboxes.length === 0 ? (
              <div className="flex min-h-[320px] flex-col items-center justify-center gap-3 text-center">
                <span className="flex size-12 items-center justify-center rounded-full bg-black/[0.04] dark:bg-white/[0.05]">
                  <Inbox className="size-5 text-black/35 dark:text-white/30" />
                </span>

                <p className="text-[13px] text-black/35 dark:text-white/30">
                  No mailboxes yet. Create one to see it here.
                </p>
              </div>
            ) : (
              <div className="space-y-2">
                {mailboxes.map((mailbox) => {
                  const address = mailbox.address
                  const username = address.split("@")[0] || address

                  return (
                    <div
                      key={address}
                      className="flex items-center justify-between gap-3 rounded-2xl border border-black/[0.08] bg-white px-4 py-4.5 shadow-[0_1px_2px_rgba(0,0,0,0.03)] transition-colors hover:border-black/[0.14] dark:border-white/[0.08] dark:bg-[#111313] dark:hover:border-white/[0.16] sm:px-5 sm:py-5"
                    >
                      <Link
                        href={`/developer/mailbox/${username}`}
                        className="group/address min-w-0 flex-1"
                      >
                        <p className="truncate font-mono text-[16px] font-medium text-black/80 transition-colors group-hover/address:text-black dark:text-white/80 dark:group-hover/address:text-white">
                          {address.includes("@")
                            ? address
                            : `${address}@${MAIL_DOMAIN}`}
                        </p>

                        <p className="mt-1.5 truncate text-[13px] text-black/35 dark:text-white/30">
                          Created{" "}
                          <span className="font-medium text-black/50 dark:text-white/45">
                            {formatDate(mailbox.createdAt)}
                          </span>
                          {validExpiry(mailbox.expiresAt) && (
                            <>
                              {" "}
                              · expires{" "}
                              <span className="font-medium text-black/50 dark:text-white/45">
                                {new Date(
                                  mailbox.expiresAt
                                ).toLocaleDateString("en-US", {
                                  month: "short",
                                  day: "numeric",
                                  year: "numeric",
                                })}
                              </span>
                            </>
                          )}
                        </p>
                      </Link>

                      <div className="flex shrink-0 items-center gap-2">
                        <button
                          type="button"
                          aria-label={`Copy ${address}`}
                          title="Copy email address"
                          onClick={() => copyMailbox(address)}
                          className="flex size-9 items-center justify-center rounded-lg text-black/30 transition-colors hover:bg-black/[0.05] hover:text-black/70 dark:text-white/30 dark:hover:bg-white/[0.08] dark:hover:text-white/70"
                        >
                          {copiedAddress === address ? (
                            <Check className="size-4 text-emerald-600" />
                          ) : (
                            <Copy className="size-4" />
                          )}
                        </button>

                        <button
                          type="button"
                          aria-label={`Delete ${address}`}
                          title="Delete mailbox"
                          onClick={() => removeMailbox(address)}
                          className="flex size-9 shrink-0 items-center justify-center rounded-lg text-black/30 transition-colors hover:bg-red-500/10 hover:text-red-500 dark:text-white/30 dark:hover:text-red-500"
                        >
                          <Trash2 className="size-4" />
                        </button>
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        </div>
      </main>
    </div>
  )
}

function formatDate(value: string | null | undefined): string {
  if (!value) return "recently"

  const time = new Date(value).getTime()
  if (Number.isNaN(time) || time <= 0) return "recently"

  return new Date(value).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  })
}

function validExpiry(value: string | null | undefined): boolean {
  if (!value) return false
  const time = new Date(value).getTime()
  return !Number.isNaN(time) && time > 0
}