"use client"

import { useEffect, useState } from "react"
import { useParams } from "next/navigation"
import { toast } from "sonner"

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://localhost:3001"

function Page() {
  const params = useParams()
  const username = params.username as string
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!username) return

    const address = `${decodeURIComponent(username)}@${process.env.NEXT_PUBLIC_MAIL_DOMAIN || "temp.mail.at"}`

    fetch(`${API_BASE}/api/mailboxes/${encodeURIComponent(address)}/message`, {
      method: "POST",
    })
      .then(async (response) => {
        if (response.ok) return

        const data = await response.json().catch(() => ({}))

        if (response.status === 404) {
          const missingError = new Error(data.error || "Mailbox not found.")
          missingError.name = "MailboxNotFound"
          throw missingError
        }

        throw new Error(data.error || "Unable to open this mailbox. Please try again.")
      })
      .catch((requestError: Error) => {
        setError(requestError.message)
        toast.error("Mailbox unavailable", {
          description: requestError.message,
          duration: 3000,
        })
      })
      .finally(() => {
        setIsLoading(false)
      })
  }, [username])

  return (
    <main className="mx-auto flex min-h-[60vh] max-w-3xl items-center justify-center px-6 py-12">
      {isLoading ? (
        <p className="text-lg text-black/50 dark:text-white/50">Opening mailbox...</p>
      ) : error ? (
        <section className="max-w-lg rounded-2xl border border-amber-500/20 bg-amber-500/5 p-6 text-center">
          <h1 className="font-heading text-2xl font-semibold">Mailbox unavailable</h1>
          <p className="mt-3 text-base leading-6 text-black/60 dark:text-white/60">
            {error}
          </p>
        </section>
      ) : (
        <p className="text-lg">Mailbox: {decodeURIComponent(username)}</p>
      )}
    </main>
  )
}

export default Page