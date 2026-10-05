"use client"

import { useCallback, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { SugestaoMeta } from "../types"

export function useMetasSugestoes() {
  const [itens, setItens] = useState<SugestaoMeta[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  const buscar = useCallback(async (busca: string) => {
    setIsLoading(true)
    setError(null)
    try {
      const query = new URLSearchParams({ busca })
      const resposta = await apiClient.get<SugestaoMeta[]>(`/metas/sugestoes?${query.toString()}`)
      setItens(resposta)
      return resposta
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsLoading(false)
    }
  }, [])

  return { itens, isLoading, error, buscar }
}
