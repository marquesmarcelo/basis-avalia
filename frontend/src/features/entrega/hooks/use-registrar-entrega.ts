"use client"

import { useCallback, useRef, useState } from "react"
import type { ErroApi } from "@/lib/api-client"
import type { Entrega } from "../types"

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:3001/api/v1"

export function useRegistrarEntrega() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)
  const chaveRef = useRef<string>(globalThis.crypto.randomUUID())
  const abortRef = useRef<AbortController | null>(null)

  const novaChave = useCallback(() => {
    chaveRef.current = globalThis.crypto.randomUUID()
  }, [])

  const registrar = useCallback(
    async (itemPlanoId: string, observacao: string, arquivos: File[]): Promise<Entrega | null> => {
      setIsSubmitting(true)
      setError(null)
      const controller = new AbortController()
      abortRef.current = controller
      try {
        const form = new FormData()
        form.append("observacao", observacao)
        for (const arquivo of arquivos) form.append("arquivos", arquivo)

        const resposta = await fetch(`${BASE_URL}/itens/${itemPlanoId}/entregas`, {
          method: "POST",
          credentials: "include",
          headers: { "Idempotency-Key": chaveRef.current },
          body: form,
          signal: controller.signal,
        })
        if (!resposta.ok) {
          const corpo = await resposta.json().catch(() => null)
          const erro: ErroApi = {
            status: resposta.status,
            code: corpo?.error?.code ?? "ERRO_INTERNO",
            message: corpo?.error?.message ?? "Não foi possível enviar agora.",
            campo: corpo?.error?.campo,
          }
          throw erro
        }
        return (await resposta.json()) as Entrega
      } catch (e) {
        setError(e as ErroApi)
        return null
      } finally {
        setIsSubmitting(false)
        abortRef.current = null
      }
    },
    []
  )

  const cancelar = useCallback(() => {
    abortRef.current?.abort()
  }, [])

  return { registrar, cancelar, isSubmitting, error, novaChave }
}
