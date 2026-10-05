"use client"

import { usePathname } from "next/navigation"
import type { ReactNode } from "react"
import { useEffect, useState } from "react"
import { Drawer, DrawerContent, DrawerTitle } from "@/components/ui/drawer"
import type { ContextoDeSessao } from "@/features/auth/types"
import { usePendencias } from "@/features/entrega/hooks/use-pendencias"
import { AppFooter } from "./app-footer"
import { AppHeader } from "./app-header"
import { AppSidebar } from "./app-sidebar"
import { AtalhosDialog } from "./atalhos-dialog"
import { navParaPerfil } from "./nav-config"

function campoDeFormularioEmFoco(): boolean {
  const ativo = document.activeElement as HTMLElement | null
  if (!ativo) return false
  const tag = ativo.tagName
  return tag === "INPUT" || tag === "TEXTAREA" || ativo.isContentEditable
}

interface AppShellProps {
  eu: ContextoDeSessao
  apenasSair?: boolean
  children: ReactNode
}

export function AppShell({ eu, apenasSair = false, children }: AppShellProps) {
  const pathname = usePathname()
  const [colapsado, setColapsado] = useState(false)
  const [menuMobileAberto, setMenuMobileAberto] = useState(false)
  const [atalhosAberto, setAtalhosAberto] = useState(false)
  const temMenu = !apenasSair && navParaPerfil(eu.permissoes).length > 0
  const { data: pendencias } = usePendencias(eu.permissoes)
  const badges = pendencias
    ? {
        pendentes_de_avaliacao: pendencias.pendentes_de_avaliacao,
        pendencias_nao_vistas: pendencias.pendencias_nao_vistas,
      }
    : undefined

  useEffect(() => {
    setMenuMobileAberto(false)
  }, [pathname])

  useEffect(() => {
    function handler(e: KeyboardEvent) {
      if (e.key === "?" && !campoDeFormularioEmFoco()) {
        setAtalhosAberto(true)
      }
    }
    document.addEventListener("keydown", handler)
    return () => document.removeEventListener("keydown", handler)
  }, [])

  return (
    <div className="flex min-h-screen flex-1 flex-col">
      <AppHeader
        eu={eu}
        apenasSair={apenasSair}
        temMenu={temMenu}
        onAlternarMenu={() => setMenuMobileAberto((v) => !v)}
        onAlternarColapso={() => setColapsado((v) => !v)}
      />
      <div className="flex flex-1">
        {temMenu && (
          <aside
            className="hidden shrink-0 border-r bg-background xl:block"
            style={{ width: colapsado ? 56 : 220 }}
          >
            <AppSidebar permissoes={eu.permissoes} colapsado={colapsado} badges={badges} />
          </aside>
        )}
        <main className="w-full flex-1 px-4 py-6 md:px-6">{children}</main>
      </div>
      <AppFooter eu={eu} />

      {temMenu && (
        <Drawer open={menuMobileAberto} onOpenChange={setMenuMobileAberto} swipeDirection="left">
          <DrawerContent className="xl:hidden">
            <DrawerTitle className="sr-only">Menu de navegação</DrawerTitle>
            <AppSidebar permissoes={eu.permissoes} badges={badges} />
          </DrawerContent>
        </Drawer>
      )}

      <AtalhosDialog open={atalhosAberto} onOpenChange={setAtalhosAberto} />
    </div>
  )
}
