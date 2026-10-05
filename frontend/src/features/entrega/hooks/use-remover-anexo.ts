"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"

export function useRemoverAnexo() {
  const [idEmAndamento, setIdEmAndamento] = useState<string | null>(null)

  async function remover(anexoId: string): Promise<boolean> {
    setIdEmAndamento(anexoId)
    try {
      await apiClient.delete(`/anexos/${anexoId}`)
      return true
    } catch (e) {
      void (e as ErroApi)
      return false
    } finally {
      setIdEmAndamento(null)
    }
  }

  return { remover, idEmAndamento }
}
