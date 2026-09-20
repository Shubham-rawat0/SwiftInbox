"use client"

import { useParams } from "next/navigation"
import { DeveloperSessionGate } from "@/components/developer/DeveloperSessionGate"
import MailboxView from "@/components/mailbox/MailboxView"
import { fetchMessages } from "@/lib/api"

const MAIL_DOMAIN =
  process.env.NEXT_PUBLIC_MAIL_DOMAIN || "temp.mail.at"

export default function DeveloperMailboxPage() {
  const { username } = useParams() as { username: string }

  const address = username.includes("@")
    ? username
    : `${username}@${MAIL_DOMAIN}`

  return (
    <div className="min-h-[calc(100svh-4rem)] bg-[#f8f8f6] text-[#171717] dark:bg-[#0b0c0c] dark:text-white">
      <main className="mx-auto w-full max-w-4xl px-5 py-6 sm:px-8 sm:py-8">
        <DeveloperSessionGate>
          <div className="animate-in fade-in duration-300 ease-out">
            <div className="mb-6 flex flex-wrap items-baseline justify-between gap-2">
              <h1 className="font-heading text-2xl font-semibold tracking-[-0.035em] sm:text-3xl">
                Mailbox
              </h1>

              <p className="truncate text-[14px] leading-5 text-black/45 dark:text-white/40">
                {address}
              </p>
            </div>

            <MailboxView
              address={address}
              basePath="/developer/mailbox"
              homeHref="/developer/mailbox"
              fetchMessages={(forceRefresh = false) =>
                fetchMessages(address, forceRefresh, true)
              }
            />
          </div>
        </DeveloperSessionGate>
      </main>
    </div>
  )
}