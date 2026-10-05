"use client"

import { useCallback, useEffect, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Usuario } from "../types"

export function useUsuario(basePath: string, id: string) {
  const [data, setData] = useState<Usuario | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<ErroApi | null>(null)

  const carregar = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const resposta = await apiClient.get<Usuario>(`${basePath}/${id}`)
      setData(resposta)
    } catch (e) {
      setError(e as ErroApi)
    } finally {
      setIsLoading(false)
    }
  }, [basePath, id])

  useEffect(() => {
    carregar()
  }, [carregar])

  return { data, isLoading, error, recarregar: carregar }
}
