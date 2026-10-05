"use client"

import { useState } from "react"
import type { ErroApi } from "@/lib/api-client"
import { concluirCarregamento, iniciarCarregamento } from "@/lib/carregamento"

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:3001/api/v1"

function nomeDoArquivo(resposta: Response, fallback: string): string {
  const cabecalho = resposta.headers.get("Content-Disposition") ?? ""
  const casado = /filename="([^"]+)"/.exec(cabecalho)
  return casado?.[1] ?? fallback
}

export function useGerarDocumento(modo: "planos" | "meus-planos") {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function gerar(planoId: string): Promise<boolean> {
    setIsSubmitting(true)
    setError(null)
    iniciarCarregamento()
    try {
      const resposta = await fetch(`${BASE_URL}/${modo}/${planoId}/documentos`, {
        method: "POST",
        credentials: "include",
      })
      if (!resposta.ok) {
        const corpo = await resposta.json().catch(() => null)
        throw {
          status: resposta.status,
          code: corpo?.error?.code ?? "ERRO_INTERNO",
          message: corpo?.error?.message ?? "Não foi possível gerar o documento agora.",
        } as ErroApi
      }
      const nome = nomeDoArquivo(resposta, "plano-de-acao.docx")
      const blob = await resposta.blob()
      const url = URL.createObjectURL(blob)
      const link = document.createElement("a")
      link.href = url
      link.download = nome
      document.body.appendChild(link)
      link.click()
      link.remove()
      URL.revokeObjectURL(url)
      return true
    } catch (e) {
      setError(e as ErroApi)
      return false
    } finally {
      setIsSubmitting(false)
      concluirCarregamento()
    }
  }

  return { gerar, isSubmitting, error, limparErro: () => setError(null) }
}
