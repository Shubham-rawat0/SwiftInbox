"use client";

import { useEffect, useRef, useState } from "react";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://localhost:3001";

export default function DocsIframe() {
  const iframeRef = useRef<HTMLIFrameElement>(null);
  const [height, setHeight] = useState<number | null>(null);

  useEffect(() => {
    const onMessage = (event: MessageEvent) => {
      if (event.source !== iframeRef.current?.contentWindow) return;

      const data = event.data as { type?: string; height?: unknown };
      if (data?.type !== "swagger:height") return;

      const raw = Number(data.height);
      if (!Number.isFinite(raw) || raw <= 0) return;

      // Small buffer so the iframe never grows its own scrollbar.
      const next = Math.round(raw) + 16;
      setHeight((prev) => (prev === next ? prev : next));
    };

    window.addEventListener("message", onMessage);
    return () => window.removeEventListener("message", onMessage);
  }, []);

  return (
    <iframe
      ref={iframeRef}
      src={`${API_BASE}/docs/`}
      className="mt-6 w-full min-w-0"
      style={{ height: height ?? "100vh" }}
    />
  );
}