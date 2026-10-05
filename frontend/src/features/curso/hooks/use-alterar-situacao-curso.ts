"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Curso, SituacaoCurso } from "../types"

export function useAlterarSituacaoCurso() {
  const [idEmAndamento, setIdEmAndamento] = useState<string | null>(null)
  const [error, setError] = useState<ErroApi | null>(null)

  async function alterarSituacao(
    id: string,
    situacao: SituacaoCurso,
    versao: number
  ): Promise<Curso | null> {
    setIdEmAndamento(id)
    setError(null)
    try {
      return await apiClient.patch<Curso>(`/cursos/${id}/situacao`, { situacao, versao })
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIdEmAndamento(null)
    }
  }

  return { alterarSituacao, idEmAndamento, error }
}
