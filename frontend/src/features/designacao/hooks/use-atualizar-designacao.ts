"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Designacao, DesignacaoInput } from "../types"

export function useAtualizarDesignacao() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function atualizar(
    id: string,
    input: DesignacaoInput,
    versao: number
  ): Promise<Designacao | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.put<Designacao>(`/designacoes/${id}`, {
        ...input,
        data_fim: input.data_fim || null,
        versao,
      })
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { atualizar, isSubmitting, error, limparErro: () => setError(null) }
}
