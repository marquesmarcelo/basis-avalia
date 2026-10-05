"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Indicador, IndicadorInput } from "../types"

export function useCriarIndicador() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function criar(input: IndicadorInput): Promise<Indicador | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.post<Indicador>("/indicadores", input)
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { criar, isSubmitting, error, limparErro: () => setError(null) }
}
