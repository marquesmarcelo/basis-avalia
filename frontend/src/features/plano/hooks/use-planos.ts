"use client"

import { useCallback, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { ParametrosListagem, RespostaPaginada } from "@/lib/tipos-api"
import type { FiltroPlanos, Plano } from "../types"

export function usePlanos(modo: "planos" | "meus-planos") {
  const [data, setData] = useState<RespostaPaginada<Plano> | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  const pesquisar = useCallback(
    async (filtros: FiltroPlanos, paginacao: ParametrosListagem) => {
      setIsLoading(true)
      setError(null)
      try {
        const query = new URLSearchParams({
          situacao: filtros.situacao,
          aprovacao: filtros.aprovacao,
          page: String(paginacao.page),
          page_size: String(paginacao.page_size),
          sort: paginacao.sort,
          order: paginacao.order,
        })
        if (filtros.periodo_id) query.set("periodo_id", filtros.periodo_id)
        if (filtros.curso_id) query.set("curso_id", filtros.curso_id)
        const resposta = await apiClient.get<RespostaPaginada<Plano>>(
          `/${modo}?${query.toString()}`
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
    [modo]
  )

  return { data, isLoading, error, pesquisar }
}
