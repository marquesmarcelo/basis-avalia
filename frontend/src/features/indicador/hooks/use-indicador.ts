"use client"

import { useCallback, useEffect, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Indicador } from "../types"

export function useIndicador(id: string) {
  const [data, setData] = useState<Indicador | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<ErroApi | null>(null)

  const carregar = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const resposta = await apiClient.get<Indicador>(`/indicadores/${id}`)
      setData(resposta)
    } catch (e) {
      setError(e as ErroApi)
    } finally {
      setIsLoading(false)
    }
  }, [id])

  useEffect(() => {
    carregar()
  }, [carregar])

  return { data, isLoading, error, recarregar: carregar }
}
