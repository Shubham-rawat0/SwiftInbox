"use client"

import { Button } from "@/components/ui/button";
import { fetchAttachment, fetchMessage, isApiError } from "@/lib/api";
import { sanitizeEmailHtml } from "@/lib/sanitize";
import type { AttachmentResult, MessageAttachment, MessageDetail } from "@/lib/types";
import { ArrowLeft, Download, FileIcon, FileText, MailX } from "lucide-react";
import Link from "next/link";
import { useEffect, useMemo, useRef, useState } from "react";
import { toast } from "sonner";
import type { AttachmentFetch } from "@/components/mailbox/AttachmentView";

const MAIL_DOMAIN = process.env.NEXT_PUBLIC_MAIL_DOMAIN || "temp.mail.at";

export type MessageDetailFetch = (messageId: string) => Promise<MessageDetail>;

const defaultFetchMessage: MessageDetailFetch = (messageId) => fetchMessage(messageId);

const defaultFetchAttachment: AttachmentFetch = (messageId, index) => fetchAttachment(messageId, index);

function formatFileSize(bytes: number): string {
    if (!bytes || bytes <= 0) return "";
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function isPreviewable(attachment: MessageAttachment): boolean {
    return (
        attachment.contentType.startsWith("image/") ||
        attachment.contentType === "application/pdf" ||
        attachment.contentType.startsWith("text/")
    );
}

export type MessageDetailViewProps = {
  username: string;
  messageId: string;
  /** Where the back button navigates (mode-specific). */
  backHref: string;
  /**
   * Bound fetcher for this access mode. Routes supply the appropriate
   * request mechanism (public vs. authenticated developer session); when
   * omitted the unauthenticated public API is used.
   */
  fetchMessage?: MessageDetailFetch;
  /**
   * Bound fetcher used to load attachment previews/downloads. Follows the
   * same access-mode rule as fetchMessage.
   */
  fetchAttachment?: AttachmentFetch;
};

function AttachmentCard({
  attachment,
  previewUrl,
  viewHref,
  onDownload,
  downloading,
}: {
  attachment: MessageAttachment;
  previewUrl?: string;
  viewHref: string;
  onDownload: () => void;
  downloading: boolean;
}) {
  const isImage = attachment.contentType.startsWith("image/");
  const isPdf = attachment.contentType === "application/pdf";
  const isText = attachment.contentType.startsWith("text/");

  return (
    <div className="flex flex-col gap-3 rounded-xl border border-gray-200 bg-gray-50/60 p-3 sm:flex-row sm:items-center dark:border-white/10 dark:bg-white/[0.04]">
      <div className="flex shrink-0 items-center gap-3 sm:flex-col sm:gap-1">
        {isImage && previewUrl ? (
          <img
            src={previewUrl}
            alt={attachment.filename || `attachment-${attachment.index}`}
            className="h-16 w-16 rounded-lg border border-gray-200 object-cover dark:border-white/10"
          />
        ) : (
          <div className="flex h-16 w-16 items-center justify-center rounded-lg border border-gray-200 bg-white dark:border-white/10 dark:bg-white/[0.04]">
            {isPdf || isText ? (
              <FileText className="h-7 w-7 text-gray-400 dark:text-gray-500" />
            ) : (
              <FileIcon className="h-7 w-7 text-gray-400 dark:text-gray-500" />
            )}
          </div>
        )}
      </div>

      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium text-gray-900 dark:text-white">
          {attachment.filename || `Attachment ${attachment.index}`}
        </p>
        <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
          {attachment.contentType}
          {attachment.size > 0 && ` · ${formatFileSize(attachment.size)}`}
          {attachment.inline && " · inline"}
        </p>
      </div>

      <div className="flex shrink-0 items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={onDownload}
          disabled={downloading}
          className="gap-1.5"
        >
          <Download className="h-3.5 w-3.5" />
          Download
        </Button>
        {isPdf && previewUrl && (
          <Link href={previewUrl} target="_blank" rel="noopener noreferrer">
            <Button variant="ghost" size="sm">View</Button>
          </Link>
        )}
        <Link href={viewHref}>
          <Button variant="ghost" size="sm">Details</Button>
        </Link>
      </div>
    </div>
  );
}

export default function MessageDetailView({
  username,
  messageId,
  backHref,
  fetchMessage: fetchMessageProp,
  fetchAttachment: fetchAttachmentProp,
}: MessageDetailViewProps) {
  const fetchDetail = fetchMessageProp ?? defaultFetchMessage;
  const fetchAttachmentFile = fetchAttachmentProp ?? defaultFetchAttachment;

  const [message, setMessage] = useState<MessageDetail | null>(null);
  const [loading, setLoading] = useState(false);
  const [attachmentUrls, setAttachmentUrls] = useState<Record<string, string>>({});
  const [downloadingIndex, setDownloadingIndex] = useState<number | null>(null);
  const objectUrlsRef = useRef<string[]>([]);
  const blobCacheRef = useRef<Map<number, AttachmentResult>>(new Map());

  useEffect(() => {
    window.scrollTo({ top: 0, left: 0, behavior: "auto" });
    document.documentElement.scrollTop = 0;
    document.body.scrollTop = 0;
  }, [messageId, username]);

  useEffect(() => {
    const urls = objectUrlsRef.current;
    return () => {
      urls.forEach((url) => URL.revokeObjectURL(url));
    };
  }, []);

  const sanitizedHtml = useMemo(() => {
    let html: string;

    if (message?.parsedData?.html) {
      html = message?.parsedData.html;
    } else if (message?.parsedData?.text) {
      html = `<p>${message.parsedData.text}</p>`;
    } else {
      html = '<p>No content available</p>';
    }

    // Resolve inline images referenced by cid: so they render without download.
    for (const att of message?.parsedData?.attachments ?? []) {
      const url = att.contentId ? attachmentUrls[`cid:${att.contentId}`] : undefined;
      if (url) {
        html = html.replace(new RegExp(`cid:${escapeRegExp(att.contentId)}`, "g"), url);
      }
    }

    return sanitizeEmailHtml(html);
  }, [message, attachmentUrls]);

  useEffect(() => {
    let isActive = true;

    const loadMessage = async () => {
      setLoading(true);
      try {
        const message = await fetchDetail(messageId);
        if (!isActive) return;
        setMessage(message);

        // Eagerly load previewable attachments (images/pdf/text) so they render inline.
        const previewable = (message.parsedData?.attachments ?? []).filter(isPreviewable);
        const urls: Record<string, string> = {};

        for (const att of previewable) {
          try {
            const result = await fetchAttachmentFile(messageId, att.index);
            const blobUrl = URL.createObjectURL(result.blob);
            objectUrlsRef.current.push(blobUrl);
            (urls as Record<string, string>)[att.contentId ? `cid:${att.contentId}` : `att-${att.index}`] = blobUrl;
            blobCacheRef.current.set(att.index, result);
          } catch (e) {
            if (!isActive) return;
            console.error('Failed to load attachment preview:', e);
          }
        }

        if (!isActive) return;
        setAttachmentUrls(urls);
      } catch (error) {
        if (!isActive) return;

        console.error('Failed to load message from API:', error);

        if (isApiError(error) && (error.status === 404 || error.status === 403)) {
          setMessage(null);
        } else {
          toast.error("Failed to load email message. Please try again later.", {
            duration: 3000,
          });
        }
      } finally {
        if (isActive) {
          setLoading(false);
        }
      }
    };

    void loadMessage();

    return () => {
      isActive = false;
    };
  }, [messageId, username, fetchDetail, fetchAttachmentFile]);

  const downloadAttachment = async (att: MessageAttachment) => {
    setDownloadingIndex(att.index);
    try {
      let result = blobCacheRef.current.get(att.index);
      if (!result) {
        result = await fetchAttachmentFile(messageId, att.index);
        blobCacheRef.current.set(att.index, result);
      }

      const url = URL.createObjectURL(result.blob);
      const anchor = document.createElement("a");
      anchor.href = url;
      anchor.download = result.filename || att.filename || `attachment-${att.index}`;
      document.body.appendChild(anchor);
      anchor.click();
      document.body.removeChild(anchor);
      setTimeout(() => URL.revokeObjectURL(url), 1000);

      toast.success("Download started");
    } catch (error) {
      console.error("Failed to download attachment:", error);
      toast.error("Failed to download attachment. Please try again.", { duration: 3000 });
    } finally {
      setDownloadingIndex(null);
    }
  };

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
  const recipient = message.mailbox || (username.includes("@") ? username : `${username}@${MAIL_DOMAIN}`)
  const attachments = message.parsedData?.attachments ?? []

  return (
    <div className="min-h-screen bg-gray-50/80 text-gray-900 dark:bg-[#0D0E0E] dark:text-white">
      <div className="mx-auto w-full max-w-4xl px-4 py-6 sm:px-6 sm:py-10 lg:px-8">

        {/* Navigation */}
        <div className="mb-5">
          <Link href={backHref}>
            <Button
              variant="ghost"
              size="sm"
              className="-ml-2 gap-2 text-sm font-medium text-dark-600 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-white/[0.06] dark:hover:text-white"
            >
              <ArrowLeft className="h-4 w-4" />
              Back to Mailbox
            </Button>
          </Link>
        </div>

        {/* Email Card */}
        <article className="overflow-hidden rounded-2xl border border-gray-200/90 bg-white shadow-[0_8px_30px_rgba(15,23,42,0.04)] dark:border-white/10 dark:bg-[#151616] dark:shadow-none">

          {/* Email Header */}
          <header className="border-b border-gray-100 px-5 py-6 dark:border-white/10 sm:px-8 sm:py-7">

            {/* Label + Date */}
            <div className="mb-3 flex items-center justify-between gap-4">
              <span className="text-xs font-medium uppercase tracking-wide text-gray-400 dark:text-gray-500">
                Email
              </span>

              <time className="shrink-0 text-sm text-gray-400 dark:text-gray-500">
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
            <h1 className="break-words text-2xl font-semibold leading-8 tracking-tight text-gray-900 dark:text-white sm:text-[1.75rem] sm:leading-9">
              {message.subject || "(no subject)"}
            </h1>

            {/* Sender */}
            <div className="mt-7 flex min-w-0 items-center gap-3.5">

              {/* Avatar */}
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-gray-100 text-sm font-semibold text-gray-600 dark:bg-white/10 dark:text-gray-300">
                {initial}
              </div>

              {/* Sender Information */}
              <div className="min-w-0 flex-1">
                <p className="truncate text-[15px] font-medium text-gray-900 dark:text-gray-100">
                  {message.from || "Unknown sender"}
                </p>

                <p className="mt-1 truncate text-sm text-gray-500 dark:text-gray-400">
                  to {recipient}
                </p>
              </div>
            </div>
          </header>

          {/* Email Body */}
          <div className="px-5 py-7 sm:px-8 sm:py-9">
            <div
              className="prose prose-base max-w-none overflow-x-auto break-words leading-7 text-gray-700 dark:prose-invert dark:text-gray-300"
              dangerouslySetInnerHTML={{ __html: sanitizedHtml }}
            />
          </div>
        </article>

        {/* Attachments */}
        {attachments.length > 0 && (
          <section className="mt-6">
            <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-gray-400 dark:text-gray-500">
              Attachments ({attachments.length})
            </h2>
            <div className="space-y-3">
              {attachments.map((att) => (
                <AttachmentCard
                  key={att.index}
                  attachment={att}
                  previewUrl={att.contentId ? attachmentUrls[`cid:${att.contentId}`] : attachmentUrls[`att-${att.index}`]}
                  viewHref={`${backHref}/message/${messageId}/attachment/${att.index}`}
                  onDownload={() => downloadAttachment(att)}
                  downloading={downloadingIndex === att.index}
                />
              ))}
            </div>
          </section>
        )}

        {/* Footer */}
        <div className="mt-4 text-center">
          <p className="text-xs text-dark-400 dark:text-gray-600">
            End of message
          </p>
        </div>
      </div>
    </div>
  )
}

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}