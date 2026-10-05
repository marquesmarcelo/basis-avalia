"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { ResultadoCopia } from "../types"

export function useCopiarEmLote() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function copiar(
    planoId: string,
    periodoDestinoId: string,
    cursos: string[]
  ): Promise<ResultadoCopia | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.post<ResultadoCopia>(`/planos/${planoId}/copias`, {
        periodo_destino_id: periodoDestinoId,
        cursos,
      })
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { copiar, isSubmitting, error, limparErro: () => setError(null) }
}
