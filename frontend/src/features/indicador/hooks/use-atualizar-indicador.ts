"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Indicador, IndicadorInput } from "../types"

export function useAtualizarIndicador() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function atualizar(
    id: string,
    input: IndicadorInput,
    versao: number
  ): Promise<Indicador | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.put<Indicador>(`/indicadores/${id}`, { ...input, versao })
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { atualizar, isSubmitting, error, limparErro: () => setError(null) }
}
