"use client"

import { createContext, useContext, useState } from "react"

export type DeveloperProfile = { name: string; email: string }

type DeveloperProfileContextValue = {
  profile: DeveloperProfile | null
  setProfile: (profile: DeveloperProfile | null) => void
}

const DeveloperProfileContext = createContext<DeveloperProfileContextValue | null>(null)

export function DeveloperProfileProvider({ children }: { children: React.ReactNode }) {
  const [profile, setProfile] = useState<DeveloperProfile | null>(null)

  return (
    <DeveloperProfileContext.Provider value={{ profile, setProfile }}>
      {children}
    </DeveloperProfileContext.Provider>
  )
}

export function useDeveloperProfile() {
  const context = useContext(DeveloperProfileContext)
  if (!context) {
    throw new Error("useDeveloperProfile must be used within DeveloperProfileProvider")
  }
  return context
}