"use client"

import { useState } from "react"
import type { ErroApi } from "@/lib/api-client"
import type { Anexo } from "../types"

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:3001/api/v1"

export function useAdicionarAnexo() {
  const [error, setError] = useState<ErroApi | null>(null)

  async function adicionar(entregaId: string, arquivo: File): Promise<Anexo[] | null> {
    setError(null)
    try {
      const form = new FormData()
      form.append("arquivos", arquivo)
      const resposta = await fetch(`${BASE_URL}/entregas/${entregaId}/anexos`, {
        method: "POST",
        credentials: "include",
        body: form,
      })
      if (!resposta.ok) {
        const corpo = await resposta.json().catch(() => null)
        throw {
          status: resposta.status,
          code: corpo?.error?.code ?? "ERRO_INTERNO",
          message: corpo?.error?.message ?? "Não foi possível enviar agora.",
        } as ErroApi
      }
      return (await resposta.json()) as Anexo[]
    } catch (e) {
      setError(e as ErroApi)
      return null
    }
  }

  return { adicionar, error }
}
