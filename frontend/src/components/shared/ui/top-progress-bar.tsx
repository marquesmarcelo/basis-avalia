"use client"

import { cn } from "cn"
import { useEffect, useState } from "react"
import { inscreverCarregamento } from "@/lib/carregamento"

export function TopProgressBar() {
  const [ativo, setAtivo] = useState(false)

  useEffect(() => inscreverCarregamento(setAtivo), [])

  return (
    <div aria-hidden="true" className="fixed inset-x-0 top-0 z-100 h-[3px] overflow-hidden">
      <div
        className={cn(
          "h-full w-full origin-left bg-primary transition-transform duration-300 ease-out motion-reduce:transition-none",
          ativo ? "scale-x-100" : "scale-x-0"
        )}
      />
    </div>
  )
}
