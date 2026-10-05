"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Plano } from "../types"

export function useExcluirItem() {
  const [idEmAndamento, setIdEmAndamento] = useState<string | null>(null)
  const [error, setError] = useState<ErroApi | null>(null)

  async function excluir(planoId: string, itemId: string): Promise<Plano | null> {
    setIdEmAndamento(itemId)
    setError(null)
    try {
      return await apiClient.delete<Plano>(`/planos/${planoId}/itens/${itemId}`)
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIdEmAndamento(null)
    }
  }

  return { excluir, idEmAndamento, error }
}
