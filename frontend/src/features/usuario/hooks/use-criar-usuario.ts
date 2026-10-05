"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Usuario, UsuarioInput } from "../types"

export function useCriarUsuario(basePath: string) {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function criar(input: UsuarioInput): Promise<Usuario | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.post<Usuario>(basePath, input)
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { criar, isSubmitting, error, limparErro: () => setError(null) }
}
