"use client"

import { signOutDeveloper } from "@/lib/api"
import { useRouter } from "next/navigation"
import { useEffect } from "react"
import { toast } from "sonner"

export default function SignOutPage() {
  const router = useRouter()

  useEffect(() => {
    let active = true

    const signOut = async () => {
      try {
        await signOutDeveloper()
        if (active) {
          toast.success("Signed out")
          router.replace("/developer")
        }
      } catch (error) {
        if (active) {
          toast.error("Unable to sign out", {
            description: error instanceof Error ? error.message : "Please try again.",
          })
          router.replace("/developer")
        }
      }
    }

    signOut()
    return () => {
      active = false
    }
  }, [router])

  return (
    <main className="flex min-h-[calc(100vh-8rem)] items-center justify-center bg-gray-50/80 px-4 dark:bg-[#0D0E0E]">
      <p className="text-sm text-gray-500 dark:text-gray-400">Signing out...</p>
    </main>
  )
}
