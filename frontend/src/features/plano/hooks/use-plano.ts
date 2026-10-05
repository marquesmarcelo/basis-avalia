"use client"

import { useCallback, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Plano } from "../types"

export function usePlano(modo: "planos" | "meus-planos") {
  const [data, setData] = useState<Plano | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  const carregar = useCallback(
    async (id: string) => {
      setIsLoading(true)
      setError(null)
      try {
        const resposta = await apiClient.get<Plano>(`/${modo}/${id}`)
        setData(resposta)
        return resposta
      } catch (e) {
        setError(e as ErroApi)
        return null
      } finally {
        setIsLoading(false)
      }
    },
    [modo]
  )

  return { data, isLoading, error, carregar, definir: setData }
}
