"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { ContextoDeSessao } from "../types"

interface AlterarSenhaInput {
  senhaAtual: string
  senhaNova: string
}

export function useAlterarSenha() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function alterarSenha(input: AlterarSenhaInput): Promise<ContextoDeSessao | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      const resposta = await apiClient.post<ContextoDeSessao>(
        "/auth/senha",
        { senha_atual: input.senhaAtual, senha_nova: input.senhaNova },
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

  return { alterarSenha, isSubmitting, error }
}
