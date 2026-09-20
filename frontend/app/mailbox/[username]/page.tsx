"use client"

import { useParams } from "next/navigation"
import MailboxView from "@/components/mailbox/MailboxView"

const MAIL_DOMAIN = process.env.NEXT_PUBLIC_MAIL_DOMAIN || "temp.mail.at"

export default function MailboxPage() {
  const { username } = useParams() as { username: string }
  const address = username.includes("@")
    ? username
    : `${username}@${MAIL_DOMAIN}`

  return (
    <MailboxView address={address} basePath="/mailbox" />
  )
}