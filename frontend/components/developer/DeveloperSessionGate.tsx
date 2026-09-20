"use client"

import { useDeveloperSession } from "@/hooks/use-developer-session"
import Link from "next/link"
import type { DeveloperDetailsResponse } from "@/lib/types"

export type DeveloperSessionGateProps = {
  children: React.ReactNode | ((profile: DeveloperDetailsResponse) => React.ReactNode)
}

/**
 * Resolves the developer session and only renders the children when a session
 * is present. Unauthenticated visitors see a sign-in prompt; transient
 * network/server failures surface an error state with a retry instead of
 * visually logging the developer out.
 */
export function DeveloperSessionGate({ children }: DeveloperSessionGateProps) {
  const session = useDeveloperSession()

  if (session.status === "loading") {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <div className="flex items-center gap-2.5 text-sm text-muted-foreground">
          <div className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
          Verifying session
        </div>
      </div>
    )
  }

  if (session.status === "error") {
    return (
      <div className="flex min-h-[60vh] items-center justify-center px-6">
        <div className="text-center">
          <p className="text-sm font-medium">
            Couldn&apos;t verify your session
          </p>

          <p className="mt-1 text-sm text-muted-foreground">
            A temporary problem occurred while reaching the server.
          </p>

          <button
            type="button"
            onClick={session.retry}
            className="mt-6 inline-flex h-10 items-center justify-center rounded-xl border border-black/10 bg-white px-5 text-sm font-semibold text-black transition hover:bg-black/[0.03] dark:border-white/10 dark:bg-white/[0.04] dark:text-white dark:hover:bg-white/[0.07]"
          >
            Try again
          </button>
        </div>
      </div>
    )
  }

  if (session.status === "unauthenticated") {
    return (
      <div className="flex min-h-[60vh] items-center justify-center px-6">
        <div className="text-center">
          <p className="text-sm font-medium">
            Sign in to view this page
          </p>

          <p className="mt-1 text-sm text-muted-foreground">
            Your developer session is required to continue.
          </p>

          <Link
            href="/signin"
            className="mt-6 inline-flex h-10 items-center justify-center rounded-xl bg-black px-5 text-sm font-semibold text-white transition hover:bg-black/80 dark:bg-white dark:text-black dark:hover:bg-white/85"
          >
            Sign in
          </Link>
        </div>
      </div>
    )
  }

  return <>{typeof children === "function" ? children(session.profile) : children}</>
}