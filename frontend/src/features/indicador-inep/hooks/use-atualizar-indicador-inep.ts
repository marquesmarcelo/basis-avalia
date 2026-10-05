"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { IndicadorInep, IndicadorInepInput } from "../types"

export function useAtualizarIndicadorInep() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function atualizar(
    id: string,
    input: IndicadorInepInput,
    versao: number
  ): Promise<IndicadorInep | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.put<IndicadorInep>(`/plataforma/indicadores/${id}`, {
        ...input,
        versao,
      })
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { atualizar, isSubmitting, error, limparErro: () => setError(null) }
}
