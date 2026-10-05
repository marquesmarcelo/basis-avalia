"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Instituicao, InstituicaoInput } from "../types"

export function useCriarInstituicao() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function criar(input: InstituicaoInput): Promise<Instituicao | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.post<Instituicao>("/instituicoes", input)
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { criar, isSubmitting, error, limparErro: () => setError(null) }
}
