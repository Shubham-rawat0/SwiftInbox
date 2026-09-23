"use client"

import { useParams } from "next/navigation"
import AttachmentView, { AttachmentFetch } from "@/components/mailbox/AttachmentView"
import { fetchAttachment } from "@/lib/api"

const MAIL_DOMAIN =
  process.env.NEXT_PUBLIC_MAIL_DOMAIN || "temp.mail.at"

const fetchDeveloperAttachment: AttachmentFetch = (messageId, index) =>
  fetchAttachment(messageId, index, true)

export default function DeveloperAttachmentPage() {
  const params = useParams()
  const username = params.username as string
  const messageId = params.messageId as string
  const index = Number(params.index)

  const address = username.includes("@")
    ? username
    : `${username}@${MAIL_DOMAIN}`

  const localUsername = address.split("@")[0] || address

  return (
    <AttachmentView
      username={address}
      messageId={messageId}
      index={index}
      backHref={`/developer/mailbox/${localUsername}/message/${messageId}`}
      fetchAttachment={fetchDeveloperAttachment}
    />
  )
}