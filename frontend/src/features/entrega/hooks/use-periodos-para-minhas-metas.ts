"use client"

import { useCallback, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"

export interface PeriodoOpcao {
  id: string
  nome: string
  data_fim: string
  aberto: boolean
}

export function usePeriodosParaMinhasMetas() {
  const [itens, setItens] = useState<PeriodoOpcao[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  const carregar = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const resposta = await apiClient.get<PeriodoOpcao[]>("/minhas-metas/periodos")
      setItens(resposta)
      return resposta
    } catch (e) {
      setError(e as ErroApi)
      return []
    } finally {
      setIsLoading(false)
    }
  }, [])

  return { itens, isLoading, error, carregar }
}
