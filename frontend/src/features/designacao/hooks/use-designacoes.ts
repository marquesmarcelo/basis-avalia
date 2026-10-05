"use client"

import { useCallback, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { ParametrosListagem, RespostaPaginada } from "@/lib/tipos-api"
import type { Designacao, FiltroDesignacoes } from "../types"

export function useDesignacoes(cursoId: string) {
  const [data, setData] = useState<RespostaPaginada<Designacao> | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  const pesquisar = useCallback(
    async (filtros: FiltroDesignacoes, paginacao: ParametrosListagem) => {
      setIsLoading(true)
      setError(null)
      try {
        const query = new URLSearchParams({
          situacao: filtros.situacao,
          page: String(paginacao.page),
          page_size: String(paginacao.page_size),
          sort: paginacao.sort,
          order: paginacao.order,
        })
        if (filtros.coordenador_id) query.set("coordenador_id", filtros.coordenador_id)
        const resposta = await apiClient.get<RespostaPaginada<Designacao>>(
          `/cursos/${cursoId}/designacoes?${query.toString()}`
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
    [cursoId]
  )

  return { data, isLoading, error, pesquisar }
}
