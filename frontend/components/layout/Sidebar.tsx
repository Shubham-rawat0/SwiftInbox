"use client"

import { ChevronsUpDown, Home, LogOut, UserRound } from "lucide-react"
import Link from "next/link"
import { useEffect, useRef, useState } from "react"
import { useDeveloperProfile } from "./DeveloperProfileContext"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import Mailbox from "@/app/developer/mailbox/page"

export function AppSidebar() {
  const [isLoggedIn, setIsLoggedIn] = useState(false)
  const [menuOpen, setMenuOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)
  const { profile } = useDeveloperProfile()

  useEffect(() => {
    const updateSession = () => {
      const loggedIn = Boolean(localStorage.getItem("developer_id"))
      setIsLoggedIn(loggedIn)
      if (!loggedIn) {
        setMenuOpen(false)
      }
    }
    updateSession()
    window.addEventListener("storage", updateSession)
    window.addEventListener("developer-session-change", updateSession)

    return () => {
      window.removeEventListener("storage", updateSession)
      window.removeEventListener("developer-session-change", updateSession)
    }
  }, [])

  // Close the account menu on outside click or Escape
  useEffect(() => {
    if (!menuOpen) return

    const onPointerDown = (event: PointerEvent) => {
      if (!menuRef.current?.contains(event.target as Node)) setMenuOpen(false)
    }
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") setMenuOpen(false)
    }

    document.addEventListener("pointerdown", onPointerDown)
    document.addEventListener("keydown", onKeyDown)
    return () => {
      document.removeEventListener("pointerdown", onPointerDown)
      document.removeEventListener("keydown", onKeyDown)
    }
  }, [menuOpen])

  const initial = profile?.name.trim().charAt(0).toUpperCase() || "D"

  return (
    <Sidebar
      collapsible="icon"
      className="top-16 h-[calc(100svh-4rem)] overflow-visible"
    >
      {/* Header: title on the left, toggle on the right (icon swaps on hover when collapsed) */}
      <SidebarHeader className="group/sidebar-header h-14 flex-row items-center justify-between px-3 py-0 group-data-[collapsible=icon]:px-2">
        <span className="truncate px-1 text-[16px] font-semibold tracking-tight group-data-[collapsible=icon]:hidden">
          Overview
        </span>
        <div className="relative size-9 shrink-0">
          <SidebarTrigger
            aria-label="Expand sidebar"
            title="Expand sidebar"
            className="absolute inset-0 hidden size-9 rounded-lg group-data-[collapsible=icon]:flex group-data-[collapsible=icon]:group-hover/sidebar-header:hidden"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true" className="size-5 fill-current">
              <path d="M5 4h14v4h-5v12h-4V8H5V4Z" />
            </svg>
          </SidebarTrigger>
          <SidebarTrigger
            aria-label="Collapse or expand sidebar"
            title="Collapse or expand sidebar"
            className="absolute inset-0 flex size-9 rounded-lg text-sidebar-foreground/70 hover:text-sidebar-foreground group-data-[collapsible=icon]:hidden group-data-[collapsible=icon]:group-hover/sidebar-header:flex"
          />
        </div>
      </SidebarHeader>

      <SidebarContent>
        <SidebarMenu className="gap-0.5 px-2 pt-1">
          <SidebarMenuItem>
            <SidebarMenuButton
              render={<Link href="/developer" />}
              isActive
              tooltip="Overview"
              className="h-9 gap-2.5 rounded-lg px-2.5 text-sm font-normal data-active:bg-sidebar-accent data-active:font-normal [&>svg]:size-[18px] [&>svg]:text-sidebar-foreground/70"
            >
              <Home/>
              <span>Home</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarContent>

      <SidebarFooter className="relative z-30 shrink-0 bg-sidebar p-2">
        {isLoggedIn ? (
          <div ref={menuRef} className="relative">
            <button
              type="button"
              aria-haspopup="menu"
              aria-expanded={menuOpen}
              onClick={() => setMenuOpen((open) => !open)}
              className="flex w-full items-center gap-2.5 rounded-lg p-2 text-left text-sm outline-none transition-colors hover:bg-sidebar-accent focus-visible:ring-2 focus-visible:ring-sidebar-ring data-[open=true]:bg-sidebar-accent group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:p-1"
              data-open={menuOpen}
            >
              <span className="flex size-8 shrink-0 items-center justify-center rounded-full bg-primary text-sm font-medium text-primary-foreground">
                {initial}
              </span>
              <span className="min-w-0 flex-1 group-data-[collapsible=icon]:hidden">
                <span className="block truncate text-sm font-medium leading-tight">
                  {profile?.name || "Developer"}
                </span>
                {profile?.email ? (
                  <span className="block truncate text-xs leading-tight text-sidebar-foreground/60">
                    {profile.email}
                  </span>
                ) : null}
              </span>
              <ChevronsUpDown className="size-4 shrink-0 text-sidebar-foreground/50 group-data-[collapsible=icon]:hidden" />
            </button>

            {menuOpen ? (
              <div
                role="menu"
                className="absolute bottom-[calc(100%+0.5rem)] left-0 z-50 w-full min-w-56 rounded-xl border bg-popover p-1.5 text-popover-foreground shadow-lg group-data-[collapsible=icon]:bottom-0 group-data-[collapsible=icon]:left-[calc(100%+0.5rem)] group-data-[collapsible=icon]:w-64"
              >
                {profile?.email ? (
                  <>
                    <div className="truncate px-2.5 py-2 text-xs text-muted-foreground">
                      {profile.email}
                    </div>
                    <div className="-mx-1.5 my-1 h-px bg-border" />
                  </>
                ) : null}
                <Link
                  href="/signout"
                  role="menuitem"
                  onClick={() => setMenuOpen(false)}
                  className="flex h-9 items-center gap-2.5 rounded-lg px-2.5 text-sm outline-none transition-colors hover:bg-accent focus-visible:bg-accent"
                >
                  <LogOut className="size-[18px] text-muted-foreground" />
                  <span>Sign out</span>
                </Link>
              </div>
            ) : null}
          </div>
        ) : (
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton
                render={<Link href="/signin" />}
                tooltip="Sign in"
                className="h-10 gap-2.5 rounded-lg px-2.5 text-sm font-medium [&>svg]:size-[18px]"
              >
                <UserRound />
                <span>Sign in</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        )}
      </SidebarFooter>
    </Sidebar>
  )
}