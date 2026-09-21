"use client"

import { Button } from "@/components/ui/button";
import { fetchAttachment, isApiError } from "@/lib/api";
import { trackEvent } from "@/lib/posthog";
import type { AttachmentResult } from "@/lib/types";
import { ArrowLeft, Download, FileX } from "lucide-react";
import Link from "next/link";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { toast } from "sonner";

export type AttachmentFetch = (messageId: string, index: number) => Promise<AttachmentResult>;

const defaultFetchAttachment: AttachmentFetch = (messageId, index) => fetchAttachment(messageId, index);

function formatFileSize(bytes: number): string {
    if (!bytes || bytes <= 0) return "";
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
}

export type AttachmentViewProps = {
    username: string;
    messageId: string;
    index: number;
    /** Where the back button navigates (mode-specific, points at the message detail). */
    backHref: string;
    /**
     * Bound fetcher for this access mode. Routes supply the appropriate
     * request mechanism (public vs. authenticated developer session); when
     * omitted the unauthenticated public API is used.
     */
    fetchAttachment?: AttachmentFetch;
};

export default function AttachmentView({
    username,
    messageId,
    index,
    backHref,
    fetchAttachment: fetchAttachmentProp,
}: AttachmentViewProps) {
    const fetchAttachmentFile = fetchAttachmentProp ?? defaultFetchAttachment;

    const [attachment, setAttachment] = useState<AttachmentResult | null>(null);
    const [loading, setLoading] = useState(true);
    const [previewUrl, setPreviewUrl] = useState<string | null>(null);
    const [previewText, setPreviewText] = useState<string>("");
    const objectUrlsRef = useRef<string[]>([]);

    useEffect(() => {
        const urls = objectUrlsRef.current;
        return () => {
            urls.forEach((url) => URL.revokeObjectURL(url));
        };
    }, []);

    useEffect(() => {
        let isActive = true;

        window.scrollTo({ top: 0, left: 0, behavior: "auto" });
        document.documentElement.scrollTop = 0;
        document.body.scrollTop = 0;

        const loadAttachment = async () => {
            setLoading(true);
            setAttachment(null);
            setPreviewUrl(null);
            setPreviewText("");
            try {
                const result = await fetchAttachmentFile(messageId, index);
                if (!isActive) return;
                setAttachment(result);

                const url = URL.createObjectURL(result.blob);
                objectUrlsRef.current.push(url);
                if (isActive) setPreviewUrl(url);

                if ((result.contentType ?? "").startsWith("text/")) {
                    result.blob.text().then((text) => {
                        if (isActive) setPreviewText(text);
                    }).catch((err) => console.error("Failed to read text preview:", err));
                }

                trackEvent("attachment_opened", {
                    username: username,
                    message_id: messageId,
                    index: index,
                    filename: result.filename,
                    content_type: result.contentType,
                    size: result.size,
                });
            } catch (error) {
                if (!isActive) return;
                const err = error instanceof Error ? error : new Error(String(error));
                trackEvent("attachment_open_failed", {
                    username: username,
                    message_id: messageId,
                    index: index,
                    error: err.message,
                });

                console.error("Failed to load attachment from API:", error);

                if (isApiError(error) && (error.status === 404 || error.status === 403)) {
                    setAttachment(null);
                } else if (!isApiError(error) || error.status !== 429) {
                    toast.error("Failed to load attachment. Please try again later.", {
                        duration: 3000,
                    });
                }
            } finally {
                if (isActive) {
                    setLoading(false);
                }
            }
        };

        void loadAttachment();

        return () => {
            isActive = false;
        };
    }, [messageId, index, username, fetchAttachmentFile]);

    const displayName = useMemo(
        () => attachment?.filename || `Attachment ${index}`,
        [attachment, index],
    );

    const isImage = (attachment?.contentType ?? "").startsWith("image/");
    const isPdf = attachment?.contentType === "application/pdf";
    const isText = (attachment?.contentType ?? "").startsWith("text/");
    const previewType = isImage ? "image" : isPdf ? "pdf" : isText ? "text" : null;
    const isEmbeddable = !!previewType && !!previewUrl;

    const download = useCallback(() => {
        if (!attachment) return;

        let url: string | null = previewUrl;
        if (!url) {
            const tmp = URL.createObjectURL(attachment.blob);
            setTimeout(() => URL.revokeObjectURL(tmp), 1000);
            url = tmp;
        }

        const anchor = document.createElement("a");
        anchor.href = url;
        anchor.download = attachment.filename || `attachment-${index}`;
        document.body.appendChild(anchor);
        anchor.click();
        document.body.removeChild(anchor);

        trackEvent("attachment_downloaded", {
            username: username,
            message_id: messageId,
            index: index,
            filename: attachment.filename,
        });
        toast.success("Download started");
    }, [attachment, previewUrl, username, messageId, index]);

    const openInNewTab = useCallback(() => {
        if (!previewUrl) return;
        window.open(previewUrl, "_blank", "noopener,noreferrer");
    }, [previewUrl]);

    if (loading) {
        return (
            <div className="min-h-screen bg-gray-100 dark:bg-[#0D0E0E] relative overflow-y-auto">
                <div className="hidden md:block">
                    <div className="mx-[60px] bg-white dark:bg-[#0D0E0E] relative z-10 h-screen flex flex-col">
                        <div className="flex items-center justify-center h-full">
                            <div className="text-center">
                                <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-500 mx-auto"></div>
                                <p className="mt-4 text-gray-500 dark:text-gray-400 text-lg">
                                    Loading attachment...
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
                <div className="md:hidden">
                    <div className="flex items-center justify-center min-h-screen">
                        <div className="text-center">
                            <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-500 mx-auto"></div>
                            <p className="mt-4 text-gray-500 dark:text-gray-400 text-lg">
                                Loading attachment...
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        );
    }

    if (!attachment) {
        return (
            <div className="min-h-screen bg-gray-50 dark:bg-[#0D0E0E] flex items-center justify-center px-4">
                <div className="w-full max-w-sm text-center">
                    <div className="mx-auto mb-5 flex h-16 w-16 items-center justify-center rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-white/10 dark:bg-white/[0.04]">
                        <FileX className="h-7 w-7 text-gray-400 dark:text-gray-500" />
                    </div>

                    <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                        Attachment not found
                    </h2>

                    <p className="mt-2 text-sm leading-6 text-gray-500 dark:text-gray-400">
                        This attachment may have expired, been deleted, or is no
                        longer available.
                    </p>

                    <Link href={backHref}>
                        <Button className="mt-6">
                            <ArrowLeft className="h-4 w-4" />
                            Back to Message
                        </Button>
                    </Link>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-gray-50/80 text-gray-900 dark:bg-[#0D0E0E] dark:text-white">
            <div className="mx-auto w-full max-w-4xl px-4 py-6 sm:px-6 sm:py-10 lg:px-8">
                <div className="mb-5">
                    <Link href={backHref}>
                        <Button
                            variant="ghost"
                            size="sm"
                            className="-ml-2 gap-2 text-sm font-medium text-dark-600 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-white/[0.06] dark:hover:text-white"
                        >
                            <ArrowLeft className="h-4 w-4" />
                            Back to Message
                        </Button>
                    </Link>
                </div>

                <article className="overflow-hidden rounded-2xl border border-gray-200/90 bg-white shadow-[0_8px_30px_rgba(15,23,42,0.04)] dark:border-white/10 dark:bg-[#151616] dark:shadow-none">
                    <header className="border-b border-gray-100 px-5 py-6 dark:border-white/10 sm:px-8 sm:py-7">
                        <div className="mb-3 flex items-center justify-between gap-4">
                            <span className="text-xs font-medium uppercase tracking-wide text-gray-400 dark:text-gray-500">
                                Attachment
                            </span>
                            {attachment.size > 0 && (
                                <span className="shrink-0 text-sm text-gray-400 dark:text-gray-500">
                                    {formatFileSize(attachment.size)}
                                </span>
                            )}
                        </div>

                        <div className="flex items-start gap-4">
                            <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-gray-100 dark:bg-white/10">
                                <Download className="h-6 w-6 text-gray-500 dark:text-gray-300" />
                            </div>

                            <div className="min-w-0 flex-1">
                                <h1 className="break-words text-2xl font-semibold leading-8 tracking-tight text-gray-900 dark:text-white sm:text-[1.75rem] sm:leading-9">
                                    {displayName}
                                </h1>

                                <p className="mt-2 truncate text-sm text-gray-500 dark:text-gray-400">
                                    {attachment.contentType || "unknown type"}
                                </p>
                            </div>
                        </div>
                    </header>

                    <div className="px-5 py-7 sm:px-8 sm:py-9">
                        <div className="flex flex-wrap gap-3">
                            <Button
                                onClick={download}
                                className="gap-2"
                            >
                                <Download className="h-4 w-4" />
                                Download {displayName}
                            </Button>
                            {isEmbeddable && (
                                <Button
                                    variant="outline"
                                    onClick={openInNewTab}
                                    className="gap-2"
                                >
                                    Open in new tab
                                </Button>
                            )}
                        </div>

                        <p className="mt-4 text-xs text-dark-400 dark:text-gray-600">
                            Click download to save this attachment to your device.
                        </p>
                    </div>

                    {isEmbeddable && (
                        <div className="px-5 pb-7 sm:px-8 sm:pb-9">
                            <div className="overflow-hidden rounded-xl border border-gray-200 bg-gray-50/60 dark:border-white/10 dark:bg-white/[0.04]">
                                {previewType === "image" && (
                                    <img
                                        src={previewUrl}
                                        alt={displayName}
                                        className="mx-auto max-h-[70vh] w-auto object-contain"
                                    />
                                )}
                                {previewType === "pdf" && (
                                    <iframe
                                        src={previewUrl}
                                        title={displayName}
                                        className="h-[70vh] w-full"
                                    />
                                )}
                                {previewType === "text" && (
                                    <pre className="max-h-[70vh] overflow-auto whitespace-pre-wrap break-words p-4 text-sm text-gray-700 dark:text-gray-300">
                                        {previewText}
                                    </pre>
                                )}
                            </div>
                        </div>
                    )}
                </article>

                <div className="mt-4 text-center">
                    <p className="text-xs text-dark-400 dark:text-gray-600">
                        End of attachment
                    </p>
                </div>
            </div>
        </div>
    );
}