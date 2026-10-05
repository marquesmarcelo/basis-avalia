"use client"

import { usePathname, useRouter } from "next/navigation"
import { useEffect } from "react"
import { ROTAS_PROTEGIDAS } from "@/components/layout/nav-config"
import { notificar } from "@/components/shared/ui/notificacoes"
import { possuiAlgumaPermissao } from "@/lib/permissoes"

export function useGuardaDePermissao(permissoes: string[] | undefined) {
  const pathname = usePathname()
  const router = useRouter()

  useEffect(() => {
    if (!permissoes || !pathname) return

    const rota = ROTAS_PROTEGIDAS.filter(
      (r) => pathname === r.path || pathname.startsWith(`${r.path}/`)
    ).sort((a, b) => b.path.length - a.path.length)[0]

    if (rota && !possuiAlgumaPermissao(permissoes, rota.permissoes)) {
      notificar.aviso("Você não tem permissão para acessar esta área.")
      router.replace("/app")
    }
  }, [permissoes, pathname, router])
}
