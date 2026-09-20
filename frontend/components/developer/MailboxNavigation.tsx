"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import { useLayoutEffect, useRef, useState } from "react"

const tabs = [
  {
    label: "Mailbox",
    href: "/developer/mailbox/mailboxes",
  },
  {
    label: "Open",
    href: "/developer/mailbox/open",
  },
]

export function MailboxNavigation() {
  const pathname = usePathname()
  const tabRefs = useRef<HTMLAnchorElement[]>([])
  const [indicator, setIndicator] = useState<{
    left: number
    width: number
  } | null>(null)

  const activeIndex = Math.max(
    0,
    tabs.findIndex((tab) => pathname === tab.href)
  )

  useLayoutEffect(() => {
    const el = tabRefs.current[activeIndex]
    if (!el) {
      setIndicator(null)
      return
    }

    setIndicator({
      left: el.offsetLeft,
      width: el.offsetWidth,
    })
  }, [activeIndex, pathname])

  return (
    <nav className="relative flex items-center gap-1 border-b border-black/[0.08] px-1 dark:border-white/[0.08]">
      {tabs.map((tab, index) => {
        const active = pathname === tab.href

        return (
          <Link
            key={tab.href}
            href={tab.href}
            ref={(node) => {
              tabRefs.current[index] = node!
            }}
            aria-current={active ? "page" : undefined}
            className={`relative px-3 py-2.5 text-[13px] font-medium transition-colors duration-200 ${
              active
                ? "text-black dark:text-white"
                : "text-black/40 hover:text-black/70 dark:text-white/40 dark:hover:text-white/70"
            }`}
          >
            {tab.label}
          </Link>
        )
      })}

      <span
        aria-hidden="true"
        className="pointer-events-none absolute -bottom-px h-px rounded-full bg-black transition-all duration-300 ease-out dark:bg-white"
        style={
          indicator
            ? {
                left: indicator.left + 8,
                width: indicator.width - 16,
              }
            : { opacity: 0 }
        }
      />
    </nav>
  )
}