"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { IndicadorInep, IndicadorInepInput } from "../types"

export function useCriarIndicadorInep() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function criar(input: IndicadorInepInput): Promise<IndicadorInep | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.post<IndicadorInep>("/plataforma/indicadores", input)
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { criar, isSubmitting, error, limparErro: () => setError(null) }
}
