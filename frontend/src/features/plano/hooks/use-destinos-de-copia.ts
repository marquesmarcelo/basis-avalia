"use client"

import { useCallback, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { DestinoDeCopia } from "../types"

export function useDestinosDeCopia() {
  const [itens, setItens] = useState<DestinoDeCopia[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  const buscar = useCallback(async (planoId: string, periodoDestinoId: string, busca: string) => {
    if (!periodoDestinoId) {
      setItens([])
      return
    }
    setIsLoading(true)
    setError(null)
    try {
      const query = new URLSearchParams({ periodo_destino_id: periodoDestinoId })
      if (busca) query.set("busca", busca)
      const resposta = await apiClient.get<DestinoDeCopia[]>(
        `/planos/${planoId}/destinos-copia?${query.toString()}`
      )
      setItens(resposta)
    } catch (e) {
      setError(e as ErroApi)
    } finally {
      setIsLoading(false)
    }
  }, [])

  return { itens, isLoading, error, buscar }
}
