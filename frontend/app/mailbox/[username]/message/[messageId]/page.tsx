"use client"

import MessageDetailView from "@/components/mailbox/MessageDetailView"
import { useParams } from "next/navigation"

export default function MessagePage() {
  const params = useParams()
  const username = params.username as string
  const messageId = params.messageId as string

  return (
    <MessageDetailView
      username={username}
      messageId={messageId}
      backHref={`/mailbox/${username}`}
    />
  )
}