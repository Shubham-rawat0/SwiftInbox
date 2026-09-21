"use client"

import AttachmentView from "@/components/mailbox/AttachmentView"
import { useParams } from "next/navigation"

export default function AttachmentPage() {
  const params = useParams()
  const username = params.username as string
  const messageId = params.messageId as string
  const index = Number(params.index)

  return (
    <AttachmentView
      username={username}
      messageId={messageId}
      index={index}
      backHref={`/mailbox/${username}/message/${messageId}`}
    />
  )
}