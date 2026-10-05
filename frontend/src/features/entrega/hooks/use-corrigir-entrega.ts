"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Entrega } from "../types"

export function useCorrigirEntrega() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function corrigir(id: string, observacao: string, versao: number): Promise<Entrega | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.put<Entrega>(`/entregas/${id}`, { observacao, versao })
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { corrigir, isSubmitting, error }
}
