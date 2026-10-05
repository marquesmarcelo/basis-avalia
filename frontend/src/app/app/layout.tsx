"use client"

import { usePathname, useRouter } from "next/navigation"
import { useEffect } from "react"
import { AppShell } from "@/components/layout/app-shell"
import { AppShellSkeleton } from "@/components/layout/app-shell-skeleton"
import { useAtalhosDeTeclado } from "@/features/auth/hooks/use-atalhos-de-teclado"
import { useEu } from "@/features/auth/hooks/use-eu"
import { useGuardaDePermissao } from "@/features/auth/hooks/use-guarda-de-permissao"
import { useGuardaSenhaProvisoria } from "@/features/auth/hooks/use-guarda-senha-provisoria"
import { useSincronizacaoDePerfis } from "@/features/auth/hooks/use-sincronizacao-de-perfis"

export default function AppLayout({ children }: LayoutProps<"/app">) {
  const { data: eu, isLoading, recarregar } = useEu()
  const router = useRouter()
  const pathname = usePathname()

  useEffect(() => {
    if (!eu) return
    if (eu.senha_provisoria && pathname !== "/app/alterar-senha") {
      router.replace("/app/alterar-senha")
    }
  }, [eu, pathname, router])

  useGuardaSenhaProvisoria(!!eu?.senha_provisoria)
  useGuardaDePermissao(eu?.senha_provisoria ? undefined : eu?.permissoes)
  useAtalhosDeTeclado()
  useSincronizacaoDePerfis(eu, recarregar)

  if (isLoading || !eu) {
    return <AppShellSkeleton />
  }

  return (
    <AppShell eu={eu} apenasSair={eu.senha_provisoria}>
      {children}
    </AppShell>
  )
}
