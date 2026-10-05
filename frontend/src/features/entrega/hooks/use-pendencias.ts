"use client"

import { useCallback, useEffect, useState } from "react"
import { apiClient } from "@/lib/api-client"
import type { Pendencias } from "../types"

export function usePendencias(permissoes: string[]) {
  const [data, setData] = useState<Pendencias | null>(null)
  const elegivel =
    permissoes.includes("entrega.registrar") || permissoes.includes("entrega.avaliar")

  const recarregar = useCallback(async () => {
    if (!elegivel) return
    try {
      const resposta = await apiClient.get<Pendencias>("/metas/pendencias")
      setData(resposta)
    } catch {}
  }, [elegivel])

  useEffect(() => {
    recarregar()
  }, [recarregar])

  return { data, recarregar }
}
