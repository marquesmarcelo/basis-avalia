"use client"

import type { LucideIcon } from "lucide-react"
import Link from "next/link"
import { usePathname } from "next/navigation"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"

interface SidebarNavProps {
  label: string
  path: string
  icon?: LucideIcon
  colapsado?: boolean
  badgeCount?: number
}

export function SidebarNav({
  label,
  path,
  icon: Icon,
  colapsado = false,
  badgeCount,
}: SidebarNavProps) {
  const pathname = usePathname()
  const ativo = pathname === path || pathname?.startsWith(`${path}/`)
  const temBadge = typeof badgeCount === "number" && badgeCount > 0

  const link = (
    <Link
      href={path}
      aria-current={ativo ? "page" : undefined}
      className={`flex items-center gap-2 rounded-lg px-2.5 py-1.5 text-sm transition-colors ${
        ativo
          ? "bg-muted font-medium text-foreground"
          : "text-muted-foreground hover:bg-muted hover:text-foreground"
      } ${colapsado ? "relative justify-center" : ""}`}
    >
      {Icon && <Icon className="size-4 shrink-0" aria-hidden="true" />}
      {!colapsado && <span>{label}</span>}
      {temBadge && (
        <span
          aria-label={`${badgeCount} pendente${badgeCount === 1 ? "" : "s"}`}
          className={`flex h-5 min-w-5 items-center justify-center rounded-full bg-destructive px-1 text-[11px] font-medium text-destructive-foreground ${colapsado ? "absolute top-0 right-0" : "ml-auto"}`}
        >
          {badgeCount > 99 ? "99+" : badgeCount}
        </span>
      )}
      {!colapsado && !temBadge && <IndicadorDeNavegacao className="ml-auto size-3.5" />}
    </Link>
  )

  if (!colapsado) return link

  return (
    <Tooltip>
      <TooltipTrigger render={link} />
      <TooltipContent side="right">{label}</TooltipContent>
    </Tooltip>
  )
}
