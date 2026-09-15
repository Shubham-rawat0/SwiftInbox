"use client"

import { Moon, Sun, Star, Mail, Code2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useTheme } from "next-themes"
import { useEffect, useState } from "react"
import Link from "next/link"
import { usePathname } from "next/navigation";

export function Header() {
  const { resolvedTheme, setTheme } = useTheme()
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  const isDark = resolvedTheme === "dark"
  const pathname=usePathname()

  const toggleDarkMode = () => {
    setTheme(isDark ? "light" : "dark")
  }

  return (
    <header className="sticky top-0 z-50 h-16 min-h-[64px] border-b border-gray-200 bg-gray-50 text-gray-950 dark:border-gray-800 dark:bg-[#111313] dark:text-white">
      <div className="mx-auto flex h-full max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">

        {/* Left side */}
        <div className="flex items-center gap-6">

          {/* Logo */}
          <Link
            href="/"
            className="group flex items-center gap-2.5"
            aria-label="TempMail home"
          >
            <div className="flex size-9 items-center justify-center rounded-lg border border-gray-200 bg-white shadow-sm transition-colors group-hover:bg-gray-100 dark:border-gray-700 dark:bg-[#191b1b] dark:group-hover:bg-[#202222]">
              <Mail
                className="size-[18px] text-gray-700 dark:text-gray-200"
                strokeWidth={2}
              />
            </div>

            <span className="font-custom text-xl font-semibold tracking-tight">
              TempMail
            </span>
          </Link>

        </div>

        {/* Right side */}
        <div className="flex items-center gap-2 sm:gap-4">
            {/* Developer */}
          {pathname!=="/developer" && <Link
            href="/developer"
            className="
              flex items-center gap-1.5
              text-sm font-medium
              text-gray-600
              transition-colors
              hover:text-gray-950
              dark:text-gray-400
              dark:hover:text-white
            "
          >
            <Code2 className="size-5" />
            <span>Developer</span>
          </Link>}
          {/* Theme toggle */}
          {mounted ? (
            <button
              type="button"
              role="switch"
              aria-label="Toggle dark mode"
              aria-checked={isDark}
              onClick={toggleDarkMode}
              className={`
                relative inline-flex h-7 w-12 shrink-0
                items-center rounded-full
                border
                transition-colors
                focus-visible:outline-none
                focus-visible:ring-2
                focus-visible:ring-gray-400
                ${
                  isDark
                    ? "border-gray-700 bg-gray-700"
                    : "border-gray-300 bg-gray-200"
                }
              `}
            >
              <span
                className={`
                  flex size-5 items-center justify-center
                  rounded-full bg-white shadow-sm
                  transition-transform duration-200
                  ${
                    isDark
                      ? "translate-x-6"
                      : "translate-x-1"
                  }
                `}
              >
                {isDark ? (
                  <Sun className="size-3 text-gray-700" />
                ) : (
                  <Moon className="size-3 text-gray-700" />
                )}
              </span>
            </button>
          ) : (
            <div className="h-7 w-12 rounded-full border border-gray-300 bg-gray-200 dark:border-gray-700 dark:bg-gray-700" />
          )}

          {/* GitHub */}
          <a
            href="https://github.com/Shubham-rawat0"
            target="_blank"
            rel="noopener noreferrer"
          >
            <Button
              variant="outline"
              size="sm"
              className="
                gap-1.5
                border-gray-300
                bg-white
                hover:bg-gray-100
                dark:border-gray-700
                dark:bg-[#191b1b]
                dark:hover:bg-[#222525]
              "
            >
              <Star
                className="size-4 fill-yellow-400 text-yellow-500"
                strokeWidth={1.5}
                aria-hidden="true"
              />

              <span className="hidden sm:inline">
                Star on GitHub
              </span>
            </Button>
          </a>

        </div>
      </div>
    </header>
  )
}

