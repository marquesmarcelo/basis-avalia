"use client"

import { useState } from "react"
import type { Entrega } from "@/features/entrega/types"
import { apiClient, type ErroApi } from "@/lib/api-client"

export function useAvaliar() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function avaliar(
    id: string,
    resultado: "aceita" | "recusada",
    motivo: string,
    versao: number
  ): Promise<Entrega | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.post<Entrega>(`/entregas/${id}/avaliacao`, {
        resultado,
        motivo,
        versao,
      })
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { avaliar, isSubmitting, error }
}
