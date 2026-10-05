"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Instituicao } from "../types"

export function useAtivarInativarInstituicao() {
  const [idEmAndamento, setIdEmAndamento] = useState<string | null>(null)
  const [error, setError] = useState<ErroApi | null>(null)

  async function alterarSituacao(
    id: string,
    situacao: "ativa" | "inativa",
    versao: number
  ): Promise<Instituicao | null> {
    setIdEmAndamento(id)
    setError(null)
    try {
      return await apiClient.patch<Instituicao>(`/instituicoes/${id}/situacao`, {
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
