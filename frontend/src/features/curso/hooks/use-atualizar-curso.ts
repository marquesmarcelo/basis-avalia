"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Curso, CursoInput } from "../types"

export function useAtualizarCurso() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function atualizar(id: string, input: CursoInput, versao: number): Promise<Curso | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.put<Curso>(`/cursos/${id}`, { ...input, versao })
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { atualizar, isSubmitting, error, limparErro: () => setError(null) }
}
