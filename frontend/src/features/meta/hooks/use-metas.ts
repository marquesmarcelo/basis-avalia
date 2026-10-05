"use client"

import { useCallback, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { ParametrosListagem, RespostaPaginada } from "@/lib/tipos-api"
import type { FiltroMetas, Meta } from "../types"

export function useMetas() {
  const [data, setData] = useState<RespostaPaginada<Meta> | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  const pesquisar = useCallback(async (filtros: FiltroMetas, paginacao: ParametrosListagem) => {
    setIsLoading(true)
    setError(null)
    try {
      const query = new URLSearchParams({
        busca: filtros.busca,
        origem: filtros.origem,
        situacao: filtros.situacao,
        page: String(paginacao.page),
        page_size: String(paginacao.page_size),
        sort: paginacao.sort,
        order: paginacao.order,
      })
      if (filtros.indicador_id) query.set("indicador_id", filtros.indicador_id)
      const resposta = await apiClient.get<RespostaPaginada<Meta>>(`/metas?${query.toString()}`)
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
