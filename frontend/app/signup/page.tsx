"use client"

import { createDeveloper, signInDeveloper } from "@/lib/api"
import { Button } from "@/components/ui/button"
import { ArrowLeft, Eye, EyeOff, UserPlus } from "lucide-react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { SyntheticEvent, useState } from "react"
import { toast } from "sonner"

export default function SignUpPage() {
  const router = useRouter()
  const [name, setName] = useState("")
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [showPassword, setShowPassword] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleSubmit = async (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault()
    setIsSubmitting(true)

    try {
      await createDeveloper({ name: name.trim(), email: email.trim(), password })
      const developer = await signInDeveloper({ email: email.trim(), password })
      localStorage.setItem("developer_id", developer.id)
      window.dispatchEvent(new Event("developer-session-change"))
      toast.success("Developer account created")
      router.push("/developer")
    } catch (error) {
      toast.error("Unable to create account", {
        description: error instanceof Error ? error.message : "Please check your details and try again.",
      })
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <main className="bg-gray-50/80 px-4 py-4 dark:bg-[#0D0E0E] sm:pt-3 sm:pb-8">
      <div className="mx-auto w-full max-w-md">
        <Link href="/developer" className="mb-4 inline-flex items-center gap-2 text-sm font-medium text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white">
          <ArrowLeft className="size-4" />
          Developer area
        </Link>

        <section className="rounded-2xl border border-gray-200 bg-white p-6 shadow-sm dark:border-white/10 dark:bg-[#151616] sm:p-8">
          <div className="mb-7">
            <div className="mb-4 flex size-10 items-center justify-center rounded-xl bg-gray-100 text-gray-700 dark:bg-white/10 dark:text-gray-200">
              <UserPlus className="size-5" />
            </div>
            <h1 className="text-2xl font-semibold tracking-tight text-gray-950 dark:text-white">Create developer account</h1>
            <p className="mt-2 text-sm leading-6 text-gray-500 dark:text-gray-400">Create an account to access developer tools.</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
              Name
              <input
                type="text"
                required
                autoComplete="name"
                value={name}
                onChange={(event) => setName(event.target.value)}
                className="mt-2 h-11 w-full rounded-lg border border-gray-300 bg-white px-3 text-sm text-gray-900 outline-none transition focus:border-gray-500 focus:ring-2 focus:ring-gray-200 dark:border-gray-700 dark:bg-[#101111] dark:text-white dark:focus:border-gray-400 dark:focus:ring-white/10"
              />
            </label>

            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
              Email
              <input
                type="email"
                required
                autoComplete="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                className="mt-2 h-11 w-full rounded-lg border border-gray-300 bg-white px-3 text-sm text-gray-900 outline-none transition focus:border-gray-500 focus:ring-2 focus:ring-gray-200 dark:border-gray-700 dark:bg-[#101111] dark:text-white dark:focus:border-gray-400 dark:focus:ring-white/10"
              />
            </label>

            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
              Password
              <div className="relative mt-2">
                <input
                  type={showPassword ? "text" : "password"}
                  required
                  minLength={8}
                  autoComplete="new-password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  className="h-11 w-full rounded-lg border border-gray-300 bg-white px-3 pr-10 text-sm text-gray-900 outline-none transition focus:border-gray-500 focus:ring-2 focus:ring-gray-200 dark:border-gray-700 dark:bg-[#101111] dark:text-white dark:focus:border-gray-400 dark:focus:ring-white/10"
                />
                <button
                  type="button"
                  aria-label={showPassword ? "Hide password" : "Show password"}
                  onClick={() => setShowPassword((visible) => !visible)}
                  className="absolute inset-y-0 right-0 flex w-10 items-center justify-center text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
                >
                  {showPassword ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
                </button>
              </div>
            </label>

            <Button type="submit" disabled={isSubmitting} className="h-11 w-full">
              {isSubmitting ? "Creating account..." : "Sign up"}
            </Button>
          </form>

          <p className="mt-5 text-center text-sm text-gray-500 dark:text-gray-400">
            Already have an account?{" "}
            <Link href="/signin" className="font-medium text-gray-900 underline underline-offset-4 hover:text-gray-600 dark:text-white dark:hover:text-gray-300">
              Sign in
            </Link>
          </p>
        </section>
      </div>
    </main>
  )
}
