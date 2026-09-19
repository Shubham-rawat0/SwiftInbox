"use client"

import { getDeveloper } from "@/lib/api"
import { DeveloperDetailsResponse } from "@/lib/types"
import { useEffect, useState } from "react"
import { toast } from "sonner"

export default function DeveloperPage() {
  const [developer, setDeveloper] = useState<DeveloperDetailsResponse | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    let isActive = true

    async function fetchDeveloper() {
      try {
        const response = await getDeveloper()
        if (isActive) {
          setDeveloper(response)
        }
      } catch (error) {
        if (isActive) {
          toast.error("Couldn't get developer", {
            description: error instanceof Error ? error.message : "Please sign in again.",
          })
        }
      } finally {
        if (isActive) {
          setIsLoading(false)
        }
      }
    }

    void fetchDeveloper()

    return () => {
      isActive = false
    }
  }, [])

  if (isLoading) {
    return <div>Loading developer…</div>
  }

  return <div>{developer ? <p>{developer.Name}</p> : <p>Developer not found</p>}</div>
}