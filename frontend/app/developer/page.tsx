"use client"

import { createCustomMailbox } from "@/lib/api"
import { DeveloperSessionGate } from "@/components/developer/DeveloperSessionGate"
import { useState } from "react"
import { toast } from "sonner"

type ExpiryOption = "1d" | "1w" | "1m" | "custom"

const MAIL_DOMAIN =
  process.env.NEXT_PUBLIC_MAIL_DOMAIN || "temp.mail.at"

function expiryToDate(
  expiry: ExpiryOption,
  customExpiry: string
): Date | undefined {
  if (expiry === "1d") {
    return new Date(Date.now() + 24 * 60 * 60 * 1000)
  }

  if (expiry === "1w") {
    return new Date(Date.now() + 7 * 24 * 60 * 60 * 1000)
  }

  if (expiry === "1m") {
    return new Date(Date.now() + 30 * 24 * 60 * 60 * 1000)
  }

  if (expiry === "custom" && customExpiry) {
    return new Date(`${customExpiry}T23:59:59.999Z`)
  }

  return undefined
}

export default function DeveloperMailboxesPage() {
  return (
    <DeveloperSessionGate>
      {(developer) => <DeveloperMailboxCreator email={developer.Email} />}
    </DeveloperSessionGate>
  )
}

function DeveloperMailboxCreator({ email }: { email: string }) {
  const [username, setUsername] = useState("")
  const [expiry, setExpiry] = useState<ExpiryOption>("1d")
  const [customExpiry, setCustomExpiry] = useState("")
  const [isCreating, setIsCreating] = useState(false)

  const validateUsername = (value: string) => {
    return value
      .split("@")[0]
      .toLowerCase()
      .replace(/[^a-z0-9]/g, "")
      .slice(0, 32)
  }

  const createMailbox = async () => {
    const cleanUsername = username.trim()

    if (!cleanUsername) {
      toast.error("Enter a username")
      return
    }

    if (expiry === "custom" && !customExpiry) {
      toast.error("Choose an expiry date")
      return
    }

    setIsCreating(true)

    try {
      const response = await createCustomMailbox(
        cleanUsername,
        true,
        expiryToDate(expiry, customExpiry)
      )

      toast.success(`Created mailbox ${response.address}`)
      setUsername("")
    } catch (error) {
      toast.error("Could not create mailbox", {
        description:
          error instanceof Error
            ? error.message
            : "Something went wrong.",
      })
    } finally {
      setIsCreating(false)
    }
  }

  const emailAddress = `${username || "username"}@${MAIL_DOMAIN}`

  return (
    <div className="min-h-[calc(100svh-4rem)] bg-[#f8f8f6] text-[#171717] dark:bg-[#0b0c0c] dark:text-white">
      <main className="mx-auto w-full max-w-4xl px-5 py-6 sm:px-8 sm:py-8">
        <div className="animate-in fade-in duration-300 ease-out">
          <div className="mb-6">
            <h1 className="font-heading text-2xl font-semibold tracking-[-0.035em] sm:text-3xl">
              Create a mailbox
            </h1>

            <p className="mt-1.5 text-[14px] leading-5 text-black/45 dark:text-white/40">
              Configure a temporary mailbox for your development workflow.
            </p>
          </div>

          <section className="mt-6 overflow-hidden rounded-2xl border border-black/[0.09] bg-white shadow-[0_12px_40px_rgba(0,0,0,0.05)] dark:border-white/[0.08] dark:bg-[#111313] dark:shadow-black/20">

            {/* Section header */}
            <div className="border-b border-black/[0.07] px-5 py-4 dark:border-white/[0.07]">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-semibold">
                    New mailbox
                  </p>

                  <p className="mt-0.5 text-xs text-black/40 dark:text-white/35">
                    Choose an address and expiration.
                  </p>
                </div>

                <span className="rounded-md bg-black/[0.04] px-2 py-1 font-mono text-[10px] font-medium tracking-wide text-black/40 dark:bg-white/[0.05] dark:text-white/35">
                  TEMP
                </span>
              </div>
            </div>

            <div className="p-5 sm:p-6">

              {/* Username */}
              <div>
                <label
                  htmlFor="mailbox-username"
                  className="mb-2 block text-xs font-medium text-black/50 dark:text-white/40"
                >
                  Email address
                </label>

                <div className="flex h-12 overflow-hidden rounded-xl border border-black/[0.10] bg-[#fafaf9] transition focus-within:border-black/25 focus-within:ring-4 focus-within:ring-black/[0.035] dark:border-white/[0.10] dark:bg-[#0b0c0c] dark:focus-within:border-white/25 dark:focus-within:ring-white/[0.035]">

                  <input
                    id="mailbox-username"
                    type="text"
                    autoComplete="off"
                    value={username}
                    placeholder="yourname"
                    disabled={isCreating}
                    onChange={(e) =>
                      setUsername(validateUsername(e.target.value))
                    }
                    className="min-w-0 flex-1 bg-transparent px-3.5 text-[15px] font-medium outline-none placeholder:text-black/25 disabled:opacity-50 dark:placeholder:text-white/20"
                  />

                  <div className="flex items-center border-l border-black/[0.07] px-3 text-sm text-black/35 dark:border-white/[0.07] dark:text-white/30">
                    @{MAIL_DOMAIN}
                  </div>
                </div>
              </div>

              {/* Expiry */}
              <div className="mt-6">
                <label className="mb-2 block text-xs font-medium text-black/50 dark:text-white/40">
                  Expires after
                </label>

                <div className="grid grid-cols-4 gap-1 rounded-xl bg-black/[0.035] p-1 dark:bg-white/[0.045]">
                  {[
                    { value: "1d", label: "1 day" },
                    { value: "1w", label: "1 week" },
                    { value: "1m", label: "1 month" },
                    { value: "custom", label: "Custom" },
                  ].map((option) => {
                    const selected = expiry === option.value

                    return (
                      <button
                        key={option.value}
                        type="button"
                        onClick={() =>
                          setExpiry(
                            option.value as ExpiryOption
                          )
                        }
                        className={`h-9 rounded-lg text-xs font-medium transition ${
                          selected
                            ? "bg-white text-black shadow-sm dark:bg-[#1d2020] dark:text-white"
                            : "text-black/45 hover:text-black dark:text-white/40 dark:hover:text-white"
                        }`}
                      >
                        {option.label}
                      </button>
                    )
                  })}
                </div>

                {expiry === "custom" && (
                  <div className="mt-2">
                    <input
                      type="date"
                      value={customExpiry}
                      min={new Date()
                        .toISOString()
                        .split("T")[0]}
                      onChange={(e) =>
                        setCustomExpiry(e.target.value)
                      }
                      className="h-10 w-full rounded-lg border border-black/[0.09] bg-[#fafaf9] px-3 text-sm outline-none transition focus:border-black/20 focus:ring-4 focus:ring-black/[0.035] dark:border-white/[0.09] dark:bg-[#0b0c0c] dark:focus:border-white/20"
                    />
                  </div>
                )}
              </div>

              {/* Address preview */}
              <div className="mt-5 flex items-center justify-between rounded-xl bg-black/[0.025] px-3.5 py-3 dark:bg-white/[0.035]">
                <span className="text-xs text-black/40 dark:text-white/35">
                  Address
                </span>

                <span className="max-w-[65%] truncate font-mono text-xs font-medium text-black/60 dark:text-white/55">
                  {emailAddress}
                </span>
              </div>

              {/* Create */}
              <button
                type="button"
                disabled={isCreating || !username.trim()}
                onClick={createMailbox}
                className="mt-4 h-11 w-full rounded-xl bg-black text-sm font-semibold text-white transition hover:bg-black/80 active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-30 dark:bg-white dark:text-black dark:hover:bg-white/85"
              >
                {isCreating
                  ? "Creating mailbox..."
                  : "Create mailbox"}
              </button>
            </div>
          </section>

          {/* Developer identity */}
          <div className="mt-6 flex items-center justify-between px-1 text-xs">
            <span className="text-black/35 dark:text-white/25">
              Signed in as
            </span>

            <span className="font-mono text-black/50 dark:text-white/40">
              {email}
            </span>
          </div>
        </div>
      </main>
    </div>
  )
}