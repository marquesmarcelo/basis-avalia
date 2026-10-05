"use client"

import { useCallback, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Candidato } from "../types"

export function useCandidatosDesignacao() {
  const [itens, setItens] = useState<Candidato[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  const buscar = useCallback(async (busca: string) => {
    setIsLoading(true)
    setError(null)
    try {
      const query = new URLSearchParams({ busca })
      const resposta = await apiClient.get<Candidato[]>(
        `/designacoes/candidatos?${query.toString()}`
      )
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
