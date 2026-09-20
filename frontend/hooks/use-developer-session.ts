"use client"

import { useCallback, useEffect, useState } from "react"
import { getDeveloper, isApiError } from "@/lib/api"
import { useDeveloperProfile } from "@/components/layout/DeveloperProfileContext"
import type { DeveloperDetailsResponse } from "@/lib/types"

export type DeveloperSession =
  | { status: "loading"; profile: null }
  | { status: "authenticated"; profile: DeveloperDetailsResponse }
  | { status: "unauthenticated"; profile: null }
  | { status: "error"; profile: null; error: Error }

/**
 * Resolves the current developer session.
 *
 * Distinguishes an unauthenticated session (401 or missing stored session)
 * from a genuine network/server failure, so a transient API error does not
 * visually log the developer out. Session validity is enforced server-side
 * via the signed developer cookie; the stored developer id is only used to
 * fetch profile details.
 */
export function useDeveloperSession() {
  const { setProfile } = useDeveloperProfile()
  const [attempt, setAttempt] = useState(0)
  const [session, setSession] = useState<DeveloperSession>({
    status: "loading",
    profile: null,
  })

  const retry = useCallback(() => {
    setSession({ status: "loading", profile: null })
    setAttempt((attempt) => attempt + 1)
  }, [])

  useEffect(() => {
    let isActive = true

    getDeveloper()
      .then((profile) => {
        if (!isActive) return
        setProfile({ name: profile.Name, email: profile.Email })
        setSession({ status: "authenticated", profile })
      })
      .catch((error: unknown) => {
        if (!isActive) return
        setProfile(null)
        if (isApiError(error) && error.status === 401) {
          setSession({ status: "unauthenticated", profile: null })
        } else {
          setSession({
            status: "error",
            profile: null,
            error: error instanceof Error ? error : new Error(String(error)),
          })
        }
      })

    return () => {
      isActive = false
    }
  }, [setProfile, attempt])

  return { ...session, retry }
}