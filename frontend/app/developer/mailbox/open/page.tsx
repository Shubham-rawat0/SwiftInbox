"use client"

import { useRouter } from "next/navigation"
import { useState } from "react"
import { MailboxNavigation } from "@/components/developer/MailboxNavigation"

const MAIL_DOMAIN =
  process.env.NEXT_PUBLIC_MAIL_DOMAIN || "temp.mail.at"

export default function OpenMailboxPage() {
  const router = useRouter()
  const [username, setUsername] = useState("")

  const emailAddress = username || `username@${MAIL_DOMAIN}`

  const cleanUsername = (value: string) =>
    value
      .split("@")[0]
      .toLowerCase()
      .replace(/[^a-z0-9]/g, "")
      .slice(0, 32)

  const openMailbox = (event: React.FormEvent) => {
    event.preventDefault()

    if (!username.trim()) return

    router.push(`/mailbox/${username.trim()}`)
  }

  return (
    <div className="min-h-[calc(100svh-4rem)] bg-[#f8f8f6] text-[#171717] dark:bg-[#0b0c0c] dark:text-white">
      <main className="mx-auto w-full max-w-4xl px-5 py-6 sm:px-8 sm:py-8">
        <div className="animate-in fade-in duration-300 ease-out">
          <div className="mb-6">
            <h1 className="font-heading text-2xl font-semibold tracking-[-0.035em] sm:text-3xl">
              Open
            </h1>

            <p className="mt-1.5 text-[14px] leading-5 text-black/45 dark:text-white/40">
              Open an existing mailbox and view its messages.
            </p>
          </div>

          <MailboxNavigation />

          <div className="mx-auto max-w-2xl py-8">

            <form
              onSubmit={openMailbox}
              className="overflow-hidden rounded-2xl border border-black/[0.08] bg-white shadow-[0_8px_30px_rgba(0,0,0,0.035)] dark:border-white/[0.08] dark:bg-[#111313] dark:shadow-black/20"
            >
              <div className="border-b border-black/[0.07] px-5 py-4 dark:border-white/[0.07]">
                <p className="text-[14px] font-semibold">
                  Open mailbox
                </p>

                <p className="mt-0.5 text-[12px] text-black/40 dark:text-white/35">
                  Enter an existing mailbox address.
                </p>
              </div>

              <div className="p-5 sm:p-6">

                <label
                  htmlFor="open-mailbox"
                  className="mb-2 block text-[12px] font-medium text-black/50 dark:text-white/40"
                >
                  Email address
                </label>

                <div className="flex h-11 overflow-hidden rounded-xl border border-black/[0.10] bg-[#fafaf9] transition focus-within:border-black/25 focus-within:ring-4 focus-within:ring-black/[0.035] dark:border-white/[0.10] dark:bg-[#0b0c0c]">
                  <input
                    id="open-mailbox"
                    type="text"
                    autoComplete="off"
                    value={username}
                    placeholder="yourname"
                    onChange={(e) =>
                      setUsername(cleanUsername(e.target.value))
                    }
                    className="min-w-0 flex-1 bg-transparent px-3.5 text-[14px] font-medium outline-none placeholder:text-black/25 dark:placeholder:text-white/20"
                  />

                  <div className="flex items-center border-l border-black/[0.07] px-3 text-[13px] text-black/35 dark:border-white/[0.07] dark:text-white/30">
                    @{MAIL_DOMAIN}
                  </div>
                </div>

                <div className="mt-5 flex items-center justify-between rounded-xl bg-black/[0.025] px-3.5 py-3 dark:bg-white/[0.035]">
                  <span className="text-[12px] text-black/40 dark:text-white/35">
                    Opening
                  </span>

                  <span className="max-w-[65%] truncate font-mono text-[12px] font-medium text-black/60 dark:text-white/55">
                    {emailAddress}
                  </span>
                </div>

                <button
                  type="submit"
                  disabled={!username.trim()}
                  className="mt-4 h-11 w-full rounded-xl bg-black text-[13px] font-semibold text-white transition hover:bg-black/80 active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-30 dark:bg-white dark:text-black dark:hover:bg-white/85"
                >
                  Open mailbox
                </button>
              </div>
            </form>
          </div>
        </div>
      </main>
    </div>
  )
}