"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Designacao, DesignacaoInput } from "../types"

export function useCriarDesignacao(cursoId: string) {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function criar(input: DesignacaoInput): Promise<Designacao | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.post<Designacao>(`/cursos/${cursoId}/designacoes`, {
        ...input,
        data_fim: input.data_fim || null,
      })
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { criar, isSubmitting, error, limparErro: () => setError(null) }
}
