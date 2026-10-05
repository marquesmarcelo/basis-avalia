"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { ContextoDeSessao } from "../types"

interface LoginInput {
  instituicaoId: string | null
  email: string
  senha: string
}

export function useLogin() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function login(input: LoginInput): Promise<ContextoDeSessao | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      const resposta = await apiClient.post<ContextoDeSessao>(
        "/auth/login",
        { instituicao_id: input.instituicaoId, email: input.email, senha: input.senha },
        { ignorarInterceptor401: true }
      )
      return resposta
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { login, isSubmitting, error }
}
