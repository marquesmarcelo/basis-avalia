"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Meta, MetaInput } from "../types"

export function useAtualizarMeta() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function atualizar(id: string, input: MetaInput, versao: number): Promise<Meta | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.put<Meta>(`/metas/${id}`, { ...input, versao })
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { atualizar, isSubmitting, error, limparErro: () => setError(null) }
}
