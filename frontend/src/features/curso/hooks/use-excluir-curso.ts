"use client"

import { useState } from "react"
import { notificar } from "@/components/shared/ui/notificacoes"
import { apiClient, type ErroApi } from "@/lib/api-client"

export function useExcluirCurso() {
  const [idEmAndamento, setIdEmAndamento] = useState<string | null>(null)
  const [error, setError] = useState<ErroApi | null>(null)

  async function excluir(id: string): Promise<boolean> {
    setIdEmAndamento(id)
    setError(null)
    try {
      await apiClient.delete(`/cursos/${id}`)
      return true
    } catch (e) {
      const erro = e as ErroApi
      setError(erro)
      notificar.erro(erro.message || "Não foi possível excluir agora.")
      return false
    } finally {
      setIdEmAndamento(null)
    }
  }

  return { excluir, idEmAndamento, error }
}
