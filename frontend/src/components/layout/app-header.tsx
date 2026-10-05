"use client"

import { MenuIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import type { ContextoDeSessao } from "@/features/auth/types"
import { BadgeInstituicao } from "./badge-instituicao"
import { MenuUsuario } from "./menu-usuario"

interface AppHeaderProps {
  eu: ContextoDeSessao
  apenasSair?: boolean
  temMenu?: boolean
  onAlternarMenu?: () => void
  onAlternarColapso?: () => void
}

export function AppHeader({
  eu,
  apenasSair = false,
  temMenu = false,
  onAlternarMenu,
  onAlternarColapso,
}: AppHeaderProps) {
  return (
    <header className="flex h-14 shrink-0 items-center gap-3 border-b bg-background px-4 md:px-6">
      {temMenu && (
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label="Alternar menu"
          className="xl:hidden"
          onClick={onAlternarMenu}
        >
          <MenuIcon aria-hidden="true" />
        </Button>
      )}
      {temMenu && (
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label="Alternar menu"
          className="hidden xl:inline-flex"
          onClick={onAlternarColapso}
        >
          <MenuIcon aria-hidden="true" />
        </Button>
      )}
      <span className="shrink-0 font-heading text-base font-semibold">basis-avalia</span>
      <div className="ml-2">
        <BadgeInstituicao eu={eu} />
      </div>
      <div className="ml-auto flex items-center gap-3">
        <MenuUsuario eu={eu} apenasSair={apenasSair} />
      </div>
    </header>
  )
}
