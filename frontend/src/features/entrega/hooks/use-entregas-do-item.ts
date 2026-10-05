"use client"

import { useCallback, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { RespostaPaginada } from "@/lib/tipos-api"
import type { Entrega } from "../types"

export function useEntregasDoItem() {
  const [data, setData] = useState<RespostaPaginada<Entrega> | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  const carregar = useCallback(async (itemPlanoId: string) => {
    setIsLoading(true)
    setError(null)
    try {
      const resposta = await apiClient.get<RespostaPaginada<Entrega>>(
        `/itens/${itemPlanoId}/entregas?page_size=100`
      )
      setData(resposta)
    } catch (e) {
      setError(e as ErroApi)
    } finally {
      setIsLoading(false)
    }
  }, [])

  return { data, isLoading, error, carregar }
}
