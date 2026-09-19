"use client"

import { Home, LogOut, UserRound } from "lucide-react"
import Link from "next/link"
import { useEffect, useState } from "react"
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

export function AppSidebar() {
  const [isLoggedIn, setIsLoggedIn] = useState(false)
  const [profile, setProfile] = useState<{ name: string; email: string } | null>(null)

  useEffect(() => {
    const updateSession = () => {
      const loggedIn = Boolean(localStorage.getItem("developer_id"))
      setIsLoggedIn(loggedIn)
      if (!loggedIn) {
        setProfile(null)
      }
    }
    const updateProfile = (event: Event) => {
      const profileEvent = event as CustomEvent<{ name: string; email: string } | null>
      setProfile(profileEvent.detail)
    }

    updateSession()
    window.addEventListener("storage", updateSession)
    window.addEventListener("developer-session-change", updateSession)
    window.addEventListener("developer-profile-change", updateProfile)

    return () => {
      window.removeEventListener("storage", updateSession)
      window.removeEventListener("developer-session-change", updateSession)
      window.removeEventListener("developer-profile-change", updateProfile)
    }
  }, [])

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader className="flex-row items-center justify-between">
        <span className="truncate text-sm font-semibold group-data-[collapsible=icon]:hidden">
          Developer
        </span>
        <SidebarTrigger />
      </SidebarHeader>
      <SidebarContent>
        <SidebarMenu className="p-2">
          <SidebarMenuItem>
            <SidebarMenuButton
              render={<Link href="/developer" />}
              tooltip="Overview"
            >
              <Home />
              <span>Overview</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarContent>
      <SidebarFooter>
        {isLoggedIn ? (
          <details className="group/profile relative">
            <summary className="flex cursor-pointer list-none items-center gap-2 rounded-md p-2 text-sm hover:bg-sidebar-accent group-data-[collapsible=icon]:justify-center">
              <span className="flex size-8 shrink-0 items-center justify-center rounded-full bg-sidebar-accent font-semibold">
                {profile?.name.trim().charAt(0).toUpperCase() || "D"}
              </span>
              <span className="min-w-0 group-data-[collapsible=icon]:hidden">
                <span className="block truncate font-medium">{profile?.name || "Developer"}</span>
                <span className="block truncate text-xs text-sidebar-foreground/60">{profile?.email || ""}</span>
              </span>
            </summary>
            <div className="absolute bottom-full left-0 z-20 mb-2 w-full rounded-md border bg-sidebar p-1 shadow-md group-data-[collapsible=icon]:left-12 group-data-[collapsible=icon]:w-56">
              <Link
                href="/signout"
                className="flex items-center gap-2 rounded-sm px-2 py-2 text-sm hover:bg-sidebar-accent"
              >
                <LogOut className="size-4" />
                <span>Sign out</span>
              </Link>
            </div>
          </details>
        ) : (
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton render={<Link href="/signin" />} tooltip="Sign in">
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