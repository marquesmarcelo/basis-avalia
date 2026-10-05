"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Plano } from "../types"

export function useDespublicarPlano() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function despublicar(id: string, versao: number): Promise<Plano | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.post<Plano>(`/planos/${id}/despublicar`, { versao })
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { despublicar, isSubmitting, error, limparErro: () => setError(null) }
}
