"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Periodo, PeriodoInput } from "../types"

export function useCriarPeriodo() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function criar(input: PeriodoInput): Promise<Periodo | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.post<Periodo>("/periodos", input)
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { criar, isSubmitting, error, limparErro: () => setError(null) }
}
