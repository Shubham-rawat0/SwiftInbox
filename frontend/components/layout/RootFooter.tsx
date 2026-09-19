"use client"

import { usePathname } from "next/navigation"
import { Footer } from "./Footer"

export function RootFooter() {
  return usePathname() === "/" ? <Footer /> : null
}
