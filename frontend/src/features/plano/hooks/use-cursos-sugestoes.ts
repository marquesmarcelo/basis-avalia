"use client"

import { useCallback, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { RespostaPaginada } from "@/lib/tipos-api"
import type { CursoEmbutido } from "../types"

export function useCursosSugestoes() {
  const [itens, setItens] = useState<CursoEmbutido[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  const buscar = useCallback(async (busca: string) => {
    setIsLoading(true)
    setError(null)
    try {
      const query = new URLSearchParams({
        busca,
        situacao: "ativo",
        page: "1",
        page_size: "50",
        sort: "nome",
        order: "asc",
      })
      const resposta = await apiClient.get<RespostaPaginada<CursoEmbutido>>(
        `/cursos?${query.toString()}`
      )
      setItens(resposta.data)
    } catch (e) {
      setError(e as ErroApi)
    } finally {
      setIsLoading(false)
    }
  }, [])

  return { itens, isLoading, error, buscar }
}
