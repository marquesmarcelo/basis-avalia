"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Instituicao, InstituicaoInput } from "../types"

export function useAtualizarInstituicao() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function atualizar(
    id: string,
    input: InstituicaoInput,
    versao: number
  ): Promise<Instituicao | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.put<Instituicao>(`/instituicoes/${id}`, { ...input, versao })
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { atualizar, isSubmitting, error, limparErro: () => setError(null) }
}
