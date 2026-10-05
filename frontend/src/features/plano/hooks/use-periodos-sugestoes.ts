"use client"

import { useCallback, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { RespostaPaginada } from "@/lib/tipos-api"
import type { PeriodoEmbutido } from "../types"

export function usePeriodosSugestoes() {
  const [itens, setItens] = useState<PeriodoEmbutido[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  const buscar = useCallback(async (nome: string) => {
    setIsLoading(true)
    setError(null)
    try {
      const query = new URLSearchParams({
        nome,
        situacao: "todas",
        page: "1",
        page_size: "50",
        sort: "data_inicio",
        order: "desc",
      })
      const resposta = await apiClient.get<RespostaPaginada<PeriodoEmbutido>>(
        `/periodos?${query.toString()}`
      )
      setItens(resposta.data)
    } catch (e) {
      setError(e as ErroApi)
    } finally {
      setIsLoading(false)
    }
  }, [])

  return { itens, isLoading, error, buscar }
}
