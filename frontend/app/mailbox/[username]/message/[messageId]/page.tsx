"use client"

import { Button } from "@/components/ui/button";
import { fetchMessage } from "@/lib/api";
import { trackEvent } from "@/lib/posthog";
import { sanitizeEmailHtml } from "@/lib/sanitize";
import { MessageDetail } from "@/lib/types";
import { ArrowLeft, MailX } from "lucide-react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

export default function MessagePage(){
    const params=useParams()
    const [message,setMessage]=useState<MessageDetail | null>(null)
    const [loading,setLoading]=useState(false)

    const messageId=params.messageId as string
    const username=params.username as string

    const sanitizedHtml = useMemo(() => {
        let html;

        if (message?.parsedData?.html) {
            html = message?.parsedData.html;
        } else if (message?.parsedData?.text) {
            html = `<p>${message.parsedData.text}</p>`;
        } else {
            html = '<p>No content available</p>';
        }

        return sanitizeEmailHtml(html);
    }, [message]);


    useEffect(()=>{
        const loadMessage=async ()=>{
            setLoading(true)
            try {
                const message=await fetchMessage(messageId)
                setMessage(message)

                trackEvent('message_opened', {
                username: username,
                message_id: messageId,
                subject: message.subject,
                from: message.from,
                has_html: !!message.parsedData?.html,
                has_text: !!message.parsedData?.text
                })
            } catch (error) {
                const err = error as Error;
                trackEvent('message_open_failed', {
                username: username,
                message_id: messageId,
                error: err.message
                })

                console.error('Failed to load message from API:', error)

                toast.error("Failed to load email message. Please try again later.", {
                    duration: 3000,})
            }
            finally{
                setLoading(false)
            }
        }

        loadMessage()
    },[messageId,username])

    if (loading) {
        return (
        <div
            className="min-h-screen bg-gray-100 dark:bg-[#0D0E0E] relative overflow-y-auto"
        >
            {/* Mobile loading */}
            <div className="md:hidden">
            <div className="flex items-center justify-center min-h-screen">
                <div className="text-center">
                <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-500 mx-auto"></div>
                <p className="mt-4 text-gray-500 dark:text-gray-400 text-lg">
                    Loading message...
                </p>
                </div>
            </div>
            </div>

            {/* Desktop loading */}
            <div className="hidden md:block">
            <div className="mx-[60px] bg-white dark:bg-[#0D0E0E] relative z-10 h-screen flex flex-col">
                <div className="flex items-center justify-center h-full">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-500 mx-auto"></div>
                    <p className="mt-4 text-gray-500 dark:text-gray-400 text-lg">
                    Loading message...
                    </p>
                </div>
                </div>
            </div>
            </div>
        </div>
        )
    }

    if (!message) {
    return (
        <div className="min-h-screen bg-gray-50 dark:bg-[#0D0E0E] flex items-center justify-center px-4">
        <div className="w-full max-w-sm text-center">
            <div className="mx-auto mb-5 flex h-16 w-16 items-center justify-center rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-white/10 dark:bg-white/[0.04]">
            <MailX className="h-7 w-7 text-gray-400 dark:text-gray-500" />
            </div>

            <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            Message not found
            </h2>

            <p className="mt-2 text-sm leading-6 text-gray-500 dark:text-gray-400">
            This message may have expired, been deleted, or is no longer
            available.
            </p>

            <Link href="/">
            <Button className="mt-6">
                Go Home
            </Button>
            </Link>
        </div>
        </div>
    )
    }

    const initial = (message.from || "?").trim().charAt(0).toUpperCase()

    return (
        <div className="min-h-screen bg-gray-50 dark:bg-[#0D0E0E]">
            <div className="mx-auto w-full max-w-4xl px-4 py-5 sm:px-6 sm:py-8 lg:px-8">

            {/* Navigation */}
            <div className="mb-5">
                <Link href={`/mailbox/${username}`}>
                <Button
                    variant="ghost"
                    size="sm"
                    className="-ml-2 gap-2 text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-white/[0.06] dark:hover:text-white"
                >
                    <ArrowLeft className="h-4 w-4" />
                    Back to Mailbox
                </Button>
                </Link>
            </div>

            {/* Email Card */}
            <article className="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-white/10 dark:bg-[#151616]">

                {/* Email Header */}
                <header className="border-b border-gray-100 px-5 py-5 dark:border-white/10 sm:px-7 sm:py-6">

                {/* Label + Date */}
                <div className="mb-3 flex items-center justify-between gap-4">
                    <span className="text-xs font-medium uppercase tracking-wide text-gray-400 dark:text-gray-500">
                    Email
                    </span>

                    <time className="shrink-0 text-xs text-gray-400 dark:text-gray-500">
                    {message.createdAt
                        ? new Date(message.createdAt).toLocaleDateString("en-US", {
                            month: "short",
                            day: "numeric",
                            year: "numeric",
                        })
                        : message.createdAt}
                    </time>
                </div>

                {/* Subject */}
                <h1 className="break-words text-xl font-semibold leading-7 text-gray-900 dark:text-white sm:text-2xl sm:leading-8">
                    {message.subject || "(no subject)"}
                </h1>

                {/* Sender */}
                <div className="mt-6 flex min-w-0 items-center gap-3.5">

                    {/* Avatar */}
                    <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-gray-100 text-sm font-semibold text-gray-600 dark:bg-white/10 dark:text-gray-300">
                    {initial}
                    </div>

                    {/* Sender Information */}
                    <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium text-gray-900 dark:text-gray-100">
                        {message.from || "Unknown sender"}
                    </p>

                    <p className="mt-0.5 truncate text-xs text-gray-500 dark:text-gray-400">
                        to {username}@temp.abhi.at
                    </p>
                    </div>
                </div>
                </header>

                {/* Email Body */}
                <div className="px-5 py-6 sm:px-7 sm:py-8">
                <div
                    className="prose prose-sm max-w-none overflow-x-auto break-words text-gray-700 dark:prose-invert dark:text-gray-300"
                    dangerouslySetInnerHTML={{ __html: sanitizedHtml }}
                />
                </div>
            </article>

            {/* Footer */}
            <div className="mt-4 text-center">
                <p className="text-xs text-gray-400 dark:text-gray-600">
                End of message
                </p>
            </div>
            </div>
        </div>
    )
}