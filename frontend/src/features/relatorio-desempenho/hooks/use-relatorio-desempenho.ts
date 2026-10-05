"use client"

import { useCallback, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { MetaPaginacao, ParametrosListagem } from "@/lib/tipos-api"
import type { FiltroRelatorio, LinhaRelatorio, ResumoRelatorio } from "../types"

interface RespostaRelatorio {
  data: LinhaRelatorio[]
  meta: MetaPaginacao
  resumo: ResumoRelatorio
}

function montarQuery(filtros: FiltroRelatorio, paginacao: ParametrosListagem): URLSearchParams {
  const query = new URLSearchParams({
    periodo_id: filtros.periodo_id,
    page: String(paginacao.page),
    page_size: String(paginacao.page_size),
    sort: paginacao.sort,
    order: paginacao.order,
  })
  if (filtros.curso_id) query.set("curso_id", filtros.curso_id)
  if (filtros.meta_id) query.set("meta_id", filtros.meta_id)
  if (filtros.indicador_id) query.set("indicador_id", filtros.indicador_id)
  if (filtros.origem) query.set("origem", filtros.origem)
  if (filtros.situacao) query.set("situacao", filtros.situacao)
  if (filtros.autoavaliado) query.set("autoavaliado", filtros.autoavaliado)
  if (filtros.incluir_inativos === "true") query.set("incluir_inativos", "true")
  return query
}

export function useRelatorioDesempenho() {
  const [data, setData] = useState<RespostaRelatorio | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  const pesquisar = useCallback(async (filtros: FiltroRelatorio, paginacao: ParametrosListagem) => {
    if (!filtros.periodo_id) return null
    setIsLoading(true)
    setError(null)
    try {
      const query = montarQuery(filtros, paginacao)
      const resposta = await apiClient.get<RespostaRelatorio>(
        `/relatorios/desempenho?${query.toString()}`
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

  return { data, isLoading, error, pesquisar, montarQuery }
}
