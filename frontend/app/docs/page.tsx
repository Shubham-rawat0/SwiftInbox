import type { Metadata } from "next";

import DocsIframe from "@/components/docs/DocsIframe";

export const metadata: Metadata = {
  title: "Developer API · SwiftInbox",
  description:
    "Interactive OpenAPI 3.1 reference for the SwiftInbox External Developer API.",
};

export default function DocsPage() {
  return (
    <div className="min-h-screen bg-white text-[#14161a] dark:bg-[#131413] dark:text-[#f3f2ef]">
      <main className="mx-auto max-w-[1400px] px-6 py-10 sm:px-10 lg:px-16">
        <header className="border-b border-black/[0.08] pb-6 dark:border-white/[0.08]">
          <p
            className="text-[12px] font-medium uppercase tracking-[0.12em] text-[#4a6a52] dark:text-[#7fa588]"
            style={{ fontFamily: "var(--font-mono)" }}
          >
            SwiftInbox External Developer API
          </p>
          <h1 className="mt-3 text-[30px] font-semibold leading-tight tracking-[-0.02em] sm:text-[38px]">
            Developer API reference
          </h1>
          <p className="mt-3 max-w-[62ch] text-[15px] leading-6 text-black/55 dark:text-white/40">
            Interactive OpenAPI 3.1 documentation for the{" "}
            <code className="font-mono text-[13px] text-black/70 dark:text-white/60">
              /api/v1/*
            </code>{" "}
            endpoints. Click{" "}
            <span className="font-semibold text-black/70 dark:text-white/60">
              Authorize
            </span>{" "}
            and paste an API key to send live requests from your browser.
          </p>
        </header>

        <DocsIframe />
      </main>
    </div>
  );
}