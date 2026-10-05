"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Plano } from "../types"

export function useEncerrarPlano() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function encerrar(id: string, motivo: string, versao: number): Promise<Plano | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.post<Plano>(`/planos/${id}/encerrar`, { motivo, versao })
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { encerrar, isSubmitting, error, limparErro: () => setError(null) }
}
