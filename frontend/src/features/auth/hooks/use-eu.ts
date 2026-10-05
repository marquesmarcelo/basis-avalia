"use client"

import { useCallback, useEffect, useState } from "react"
import { apiClient, type ErroApi } from "@/lib/api-client"
import type { ContextoDeSessao } from "../types"

let sessaoAtual: ContextoDeSessao | null = null
let jaCarregou = false
const assinantes = new Set<() => void>()

function publicar() {
  for (const assinante of assinantes) {
    assinante()
  }
}

async function buscarSessao(): Promise<ContextoDeSessao> {
  const resposta = await apiClient.get<ContextoDeSessao>("/auth/eu")
  sessaoAtual = resposta
  jaCarregou = true
  publicar()
  return resposta
}

export function definirSessaoAtual(contexto: ContextoDeSessao) {
  sessaoAtual = contexto
  jaCarregou = true
  publicar()
}

export function useEu() {
  const [data, setData] = useState<ContextoDeSessao | null>(sessaoAtual)
  const [isLoading, setIsLoading] = useState(!jaCarregou)
  const [error, setError] = useState<ErroApi | null>(null)

  useEffect(() => {
    const assinante = () => setData(sessaoAtual)
    assinantes.add(assinante)
    return () => {
      assinantes.delete(assinante)
    }
  }, [])

  const carregar = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      return await buscarSessao()
    } catch (e) {
      setError(e as ErroApi)
      return null
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    if (!jaCarregou) {
      carregar()
    }
  }, [carregar])

  return { data, isLoading, error, recarregar: carregar }
}
