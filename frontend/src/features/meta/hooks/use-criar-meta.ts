"use client"

import { useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { Meta, MetaInput } from "../types"

export function useCriarMeta() {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<ErroApi | null>(null)

  async function criar(input: MetaInput): Promise<Meta | null> {
    setIsSubmitting(true)
    setError(null)
    try {
      return await apiClient.post<Meta>("/metas", input)
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsSubmitting(false)
    }
  }

  return { criar, isSubmitting, error, limparErro: () => setError(null) }
}
