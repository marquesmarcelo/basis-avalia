"use client"

import { useCallback, useState } from "react"
import type { Entrega } from "@/features/entrega/types"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { ParametrosListagem, RespostaPaginada } from "@/lib/tipos-api"

export interface FiltroFila {
  [key: string]: unknown
  periodo_id: string
  curso_id: string
  meta_id: string
  situacao: string
}

export function useFilaDeAvaliacao() {
  const [data, setData] = useState<RespostaPaginada<Entrega> | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  const pesquisar = useCallback(async (filtros: FiltroFila, paginacao: ParametrosListagem) => {
    setIsLoading(true)
    setError(null)
    try {
      const query = new URLSearchParams({
        page: String(paginacao.page),
        page_size: String(paginacao.page_size),
        sort: paginacao.sort,
        order: paginacao.order,
      })
      if (filtros.periodo_id) query.set("periodo_id", filtros.periodo_id)
      if (filtros.curso_id) query.set("curso_id", filtros.curso_id)
      if (filtros.meta_id) query.set("meta_id", filtros.meta_id)
      if (filtros.situacao) query.set("situacao", filtros.situacao)
      const resposta = await apiClient.get<RespostaPaginada<Entrega>>(
        `/avaliacoes?${query.toString()}`
      )
      setData(resposta)
      return resposta
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsLoading(false)
    }
  }, [])

  return { data, isLoading, error, pesquisar }
}
