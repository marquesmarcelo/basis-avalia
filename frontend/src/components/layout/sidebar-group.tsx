"use client"

import { ChevronDownIcon } from "lucide-react"
import { usePathname } from "next/navigation"
import { useState } from "react"
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import type { NavGroup } from "./nav-config"
import { SidebarNav } from "./sidebar-nav"

interface SidebarGroupProps {
  grupo: NavGroup
  colapsado?: boolean
  badges?: Record<string, number>
}

export function SidebarGroup({ grupo, colapsado = false, badges }: SidebarGroupProps) {
  const pathname = usePathname()
  const contemAtivo = grupo.items.some(
    (item) => pathname === item.path || pathname?.startsWith(`${item.path}/`)
  )
  const [aberto, setAberto] = useState(contemAtivo)
  const Icon = grupo.icon

  if (colapsado) {
    return (
      <div className="flex flex-col gap-1">
        <Tooltip>
          <TooltipTrigger
            render={
              <div className="flex items-center justify-center rounded-lg py-1.5 text-muted-foreground">
                <Icon className="size-4" aria-hidden="true" />
              </div>
            }
          />
          <TooltipContent side="right">{grupo.label}</TooltipContent>
        </Tooltip>
        {grupo.items.map((item) => (
          <SidebarNav
            key={item.path}
            label={item.label}
            path={item.path}
            icon={item.icon}
            colapsado
            badgeCount={item.badge ? badges?.[item.badge] : undefined}
          />
        ))}
      </div>
    )
  }

  return (
    <Collapsible open={aberto} onOpenChange={setAberto}>
      <CollapsibleTrigger className="flex w-full items-center gap-2 rounded-lg px-2.5 py-1.5 text-sm font-medium text-muted-foreground hover:bg-muted hover:text-foreground">
        <Icon className="size-4 shrink-0" aria-hidden="true" />
        <span className="flex-1 text-left">{grupo.label}</span>
        <ChevronDownIcon
          className={`size-4 transition-transform ${aberto ? "rotate-180" : ""}`}
          aria-hidden="true"
        />
      </CollapsibleTrigger>
      <CollapsibleContent>
        <div className="flex flex-col gap-1 py-1 pl-4">
          {grupo.items.map((item) => (
            <SidebarNav
              key={item.path}
              label={item.label}
              path={item.path}
              icon={item.icon}
              badgeCount={item.badge ? badges?.[item.badge] : undefined}
            />
          ))}
        </div>
      </CollapsibleContent>
    </Collapsible>
  )
}
