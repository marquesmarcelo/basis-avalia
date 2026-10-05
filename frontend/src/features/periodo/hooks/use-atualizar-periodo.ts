"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Periodo, PeriodoInput } from "../types"

export function useAtualizarPeriodo() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function atualizar(
    id: string,
    input: PeriodoInput,
    versao: number
  ): Promise<Periodo | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.put<Periodo>(`/periodos/${id}`, { ...input, versao })
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { atualizar, isSubmitting, error, limparErro: () => setError(null) }
}
