"use client"

import { useCallback, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { FiltroRelatorio, ItemDesempenhoPorCurso, MetaDesempenhoPorCurso } from "../types"

interface RespostaDesempenhoPorCurso {
  data: ItemDesempenhoPorCurso[]
  cursos_sem_coordenador: string[]
  meta: MetaDesempenhoPorCurso
}

function montarQuery(filtros: FiltroRelatorio): URLSearchParams {
  const query = new URLSearchParams({ periodo_id: filtros.periodo_id })
  if (filtros.curso_id) query.set("curso_id", filtros.curso_id)
  if (filtros.meta_id) query.set("meta_id", filtros.meta_id)
  if (filtros.indicador_id) query.set("indicador_id", filtros.indicador_id)
  if (filtros.origem) query.set("origem", filtros.origem)
  if (filtros.situacao) query.set("situacao", filtros.situacao)
  if (filtros.autoavaliado) query.set("autoavaliado", filtros.autoavaliado)
  if (filtros.incluir_inativos === "true") query.set("incluir_inativos", "true")
  return query
}

export function useDesempenhoPorCurso() {
  const [data, setData] = useState<RespostaDesempenhoPorCurso | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  const pesquisar = useCallback(async (filtros: FiltroRelatorio) => {
    if (!filtros.periodo_id) return null
    setIsLoading(true)
    setError(null)
    try {
      const query = montarQuery(filtros)
      const resposta = await apiClient.get<RespostaDesempenhoPorCurso>(
        `/relatorios/desempenho/por-curso?${query.toString()}`
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
