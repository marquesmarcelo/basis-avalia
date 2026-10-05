"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"

export function useRedefinirSenha(basePath: string) {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function redefinir(id: string, senhaNova: string): Promise<boolean> {
    setIsSubmitting(true)
    setError(null)
    try {
      await apiClient.post(`${basePath}/${id}/senha`, { senha_nova: senhaNova })
      return true
    } catch (e) {
      setError(e as ErroApi)
      return false
    } finally {
      setIsSubmitting(false)
    }
  }

  return { redefinir, isSubmitting, error, limparErro: () => setError(null) }
}
