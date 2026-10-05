"use client"

import { useCallback, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { GrupoMinhasMetas } from "../types"

export function useMinhasMetas() {
  const [data, setData] = useState<GrupoMinhasMetas[] | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  const carregar = useCallback(async (periodoId: string) => {
    if (!periodoId) return
    setIsLoading(true)
    setError(null)
    try {
      const resposta = await apiClient.get<GrupoMinhasMetas[]>(
        `/minhas-metas?periodo_id=${periodoId}`
      )
      setData(resposta)
    } catch (e) {
      setError(e as ErroApi)
    } finally {
      setIsLoading(false)
    }
  }, [])

  return { data, isLoading, error, carregar }
}
