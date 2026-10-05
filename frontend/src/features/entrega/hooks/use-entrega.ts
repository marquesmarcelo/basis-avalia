"use client"

import { useCallback, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Entrega } from "../types"

export function useEntrega() {
  const [data, setData] = useState<Entrega | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  const carregar = useCallback(async (id: string) => {
    setIsLoading(true)
    setError(null)
    try {
      const resposta = await apiClient.get<Entrega>(`/entregas/${id}`)
      setData(resposta)
      return resposta
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsLoading(false)
    }
  }, [])

  return { data, isLoading, error, carregar }
}
