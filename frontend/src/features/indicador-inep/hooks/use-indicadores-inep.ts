"use client"

import { useCallback, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { ParametrosListagem, RespostaPaginada } from "@/lib/tipos-api"
import type { FiltroIndicadoresInep, IndicadorInep } from "../types"

export function useIndicadoresInep() {
  const [data, setData] = useState<RespostaPaginada<IndicadorInep> | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  const pesquisar = useCallback(
    async (filtros: FiltroIndicadoresInep, paginacao: ParametrosListagem) => {
      setIsLoading(true)
      setError(null)
      try {
        const query = new URLSearchParams({
          busca: filtros.busca,
          situacao: filtros.situacao,
          page: String(paginacao.page),
          page_size: String(paginacao.page_size),
          sort: paginacao.sort,
          order: paginacao.order,
        })
        const resposta = await apiClient.get<RespostaPaginada<IndicadorInep>>(
          `/plataforma/indicadores?${query.toString()}`
        )
        setData(resposta)
        return resposta
      } catch (e) {
        setError(e as ErroApi)
        return null
      } finally {
        setIsLoading(false)
      }
    },
    []
  )

  return { data, isLoading, error, pesquisar }
}
