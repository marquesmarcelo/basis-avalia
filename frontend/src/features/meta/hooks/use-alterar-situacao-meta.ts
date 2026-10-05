"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Meta } from "../types"

export function useAlterarSituacaoMeta() {
  const [idEmAndamento, setIdEmAndamento] = useState<string | null>(null)
  const [error, setError] = useState<ErroApi | null>(null)

  async function alterarSituacao(
    id: string,
    situacao: "ativo" | "inativo",
    versao: number
  ): Promise<Meta | null> {
    setIdEmAndamento(id)
    setError(null)
    try {
      return await apiClient.patch<Meta>(`/metas/${id}/situacao`, { situacao, versao })
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIdEmAndamento(null)
    }
  }

  return { alterarSituacao, idEmAndamento, error }
}
