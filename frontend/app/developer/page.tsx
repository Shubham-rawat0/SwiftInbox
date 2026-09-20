"use client"

import { createCustomMailbox, getDeveloper } from "@/lib/api"
import { DeveloperDetailsResponse } from "@/lib/types"
import { useDeveloperProfile } from "@/components/layout/DeveloperProfileContext"
import { useEffect, useState } from "react"
import { toast } from "sonner"

type ExpiryOption = "1d" | "1w" | "1m" | "custom"

const MAIL_DOMAIN =
  process.env.NEXT_PUBLIC_MAIL_DOMAIN || "temp.mail.at"

export default function DeveloperPage() {
  const [developer, setDeveloper] =
    useState<DeveloperDetailsResponse | null>(null)

  const [isLoading, setIsLoading] = useState(true)

  const [username, setUsername] = useState("")
  const [expiry, setExpiry] = useState<ExpiryOption>("1d")
  const [customExpiry, setCustomExpiry] = useState("")
  const [isCreating, setIsCreating] = useState(false)

  const { setProfile } = useDeveloperProfile()

  useEffect(() => {
    let isActive = true

    async function fetchDeveloper() {
      try {
        const response = await getDeveloper()

        if (!isActive) return

        setDeveloper(response)

        setProfile({
          name: response.Name,
          email: response.Email,
        })
      } catch (error) {
        if (!isActive) return

        setDeveloper(null)
        setProfile(null)

        toast.error("Developer sign-in required", {
          description:
            error instanceof Error
              ? error.message
              : "Please sign in to access the developer dashboard.",
        })
      } finally {
        if (isActive) {
          setIsLoading(false)
        }
      }
    }

    void fetchDeveloper()

    return () => {
      isActive = false
    }
  }, [setProfile])

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
      const response = await createCustomMailbox(cleanUsername,true)
      console.log("TODO: create developer mailbox", {
        username: cleanUsername,
        expiry,
        expiresAt:
          expiry === "custom"
            ? customExpiry
            : undefined,
      })

      toast.success(`Created mailbox ${response.address}`)
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

  if (isLoading) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <div className="flex items-center gap-2.5 text-sm text-muted-foreground">
          <div className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
          Loading developer workspace
        </div>
      </div>
    )
  }

  if (!developer) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center px-6">
        <div className="text-center">
          <p className="text-sm font-medium">
            Sign in to access the developer workspace
          </p>

          <p className="mt-1 text-sm text-muted-foreground">
            Your developer session is required to continue.
          </p>
        </div>
      </div>
    )
  }

  const emailAddress = `${username || "username"}@${MAIL_DOMAIN}`

  return (
    <div className="flex min-h-[calc(100svh-4rem)] flex-col bg-[#f7f7f5] text-[#111] dark:bg-[#0b0c0c] dark:text-white">
      <main className="mx-auto w-full max-w-5xl px-5 py-10 lg:px-8">
        <div className="mx-auto max-w-2xl">

            <h1 className="font-heading text-4xl font-semibold tracking-[-0.04em] sm:text-5xl">
              Create a mailbox.
            </h1>

            <p className="mt-3 max-w-lg text-[15px] leading-6 text-black/50 dark:text-white/40">
              Configure a temporary mailbox for your development
              workflow.
            </p>

          {/* Mailbox creation */}
          <section className="overflow-hidden rounded-2xl border border-black/[0.09] bg-white shadow-[0_12px_40px_rgba(0,0,0,0.05)] dark:border-white/[0.08] dark:bg-[#111313] dark:shadow-black/20">

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
              {developer.Email}
            </span>
          </div>
        </div>
      </main>
    </div>
  )
}