"use client"

import { useCallback, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { ParametrosListagem, RespostaPaginada } from "@/lib/tipos-api"
import type { FiltroUsuarios, Usuario } from "../types"

export function useUsuarios(basePath: string) {
  const [data, setData] = useState<RespostaPaginada<Usuario> | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  const pesquisar = useCallback(
    async (filtros: FiltroUsuarios, paginacao: ParametrosListagem) => {
      setIsLoading(true)
      setError(null)
      try {
        const query = new URLSearchParams({
          busca: filtros.busca,
          page: String(paginacao.page),
          page_size: String(paginacao.page_size),
          sort: paginacao.sort,
          order: paginacao.order,
        })
        if (filtros.perfil) {
          query.set("perfil", filtros.perfil)
        }
        const resposta = await apiClient.get<RespostaPaginada<Usuario>>(
          `${basePath}?${query.toString()}`
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
    [basePath]
  )

  return { data, isLoading, error, pesquisar }
}
