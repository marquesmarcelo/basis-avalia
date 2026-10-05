"use client"

import { useCallback, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { ParametrosListagem, RespostaPaginada } from "@/lib/tipos-api"
import type { Curso, FiltroCursos } from "../types"

export function useCursos() {
  const [data, setData] = useState<RespostaPaginada<Curso> | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  const pesquisar = useCallback(async (filtros: FiltroCursos, paginacao: ParametrosListagem) => {
    setIsLoading(true)
    setError(null)
    try {
      const query = new URLSearchParams({
        busca: filtros.busca,
        grau: filtros.grau === "todos" ? "" : filtros.grau,
        modalidade: filtros.modalidade === "todas" ? "" : filtros.modalidade,
        situacao: filtros.situacao,
        page: String(paginacao.page),
        page_size: String(paginacao.page_size),
        sort: paginacao.sort,
        order: paginacao.order,
      })
      if (filtros.coordenador_id) query.set("coordenador_id", filtros.coordenador_id)
      const resposta = await apiClient.get<RespostaPaginada<Curso>>(`/cursos?${query.toString()}`)
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
