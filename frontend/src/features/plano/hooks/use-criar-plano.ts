"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Plano, PlanoInput } from "../types"

export function useCriarPlano() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function criar(input: PlanoInput): Promise<Plano | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.post<Plano>("/planos", input)
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { criar, isSubmitting, error, limparErro: () => setError(null) }
}
