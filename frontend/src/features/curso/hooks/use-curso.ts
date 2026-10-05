"use client"

import { useCallback, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Curso } from "../types"

export function useCurso() {
  const [data, setData] = useState<Curso | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  const buscar = useCallback(async (id: string) => {
    setIsLoading(true)
    setError(null)
    try {
      const resposta = await apiClient.get<Curso>(`/cursos/${id}`)
      setData(resposta)
      return resposta
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsLoading(false)
    }
  }, [])

  return { data, isLoading, error, buscar }
}
