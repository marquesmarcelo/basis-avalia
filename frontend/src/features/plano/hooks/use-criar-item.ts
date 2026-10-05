"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { ItemPlanoInput, Plano } from "../types"

export function useCriarItem() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function criar(planoId: string, input: ItemPlanoInput): Promise<Plano | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.post<Plano>(`/planos/${planoId}/itens`, input)
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { criar, isSubmitting, error, limparErro: () => setError(null) }
}
