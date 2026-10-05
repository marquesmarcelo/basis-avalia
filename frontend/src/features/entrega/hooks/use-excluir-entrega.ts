"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"

export function useExcluirEntrega() {
  const [idEmAndamento, setIdEmAndamento] = useState<string | null>(null)
  const [error, setError] = useState<ErroApi | null>(null)

  async function excluir(id: string): Promise<boolean> {
    setIdEmAndamento(id)
    setError(null)
    try {
      await apiClient.delete(`/entregas/${id}`)
      return true
    } catch (e) {
      setError(e as ErroApi)
      return false
    } finally {
      setIdEmAndamento(null)
    }
  }

  return { excluir, idEmAndamento, error }
}
