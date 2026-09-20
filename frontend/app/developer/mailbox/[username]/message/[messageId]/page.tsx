"use client"

import { useParams } from "next/navigation"
import { DeveloperSessionGate } from "@/components/developer/DeveloperSessionGate"
import MessageDetailView, { MessageDetailFetch } from "@/components/mailbox/MessageDetailView"
import { fetchMessage } from "@/lib/api"

const MAIL_DOMAIN =
  process.env.NEXT_PUBLIC_MAIL_DOMAIN || "temp.mail.at"

const fetchDeveloperMessage: MessageDetailFetch = (messageId) =>
  fetchMessage(messageId, true)

export default function DeveloperMessagePage() {
  const params = useParams()
  const username = params.username as string
  const messageId = params.messageId as string

  const address = username.includes("@")
    ? username
    : `${username}@${MAIL_DOMAIN}`

  const localUsername = address.split("@")[0] || address

  return (
    <DeveloperSessionGate>
      <MessageDetailView
        username={address}
        messageId={messageId}
        backHref={`/developer/mailbox/${localUsername}`}
        fetchMessage={fetchDeveloperMessage}
      />
    </DeveloperSessionGate>
  )
}