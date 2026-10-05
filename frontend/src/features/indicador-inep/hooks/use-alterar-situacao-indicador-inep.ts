"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { IndicadorInep } from "../types"

export function useAlterarSituacaoIndicadorInep() {
  const [idEmAndamento, setIdEmAndamento] = useState<string | null>(null)
  const [error, setError] = useState<ErroApi | null>(null)

  async function alterarSituacao(
    id: string,
    situacao: "ativo" | "inativo",
    versao: number
  ): Promise<IndicadorInep | null> {
    setIdEmAndamento(id)
    setError(null)
    try {
      return await apiClient.patch<IndicadorInep>(`/plataforma/indicadores/${id}/situacao`, {
        situacao,
        versao,
      })
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIdEmAndamento(null)
    }
  }

  return { alterarSituacao, idEmAndamento, error }
}
