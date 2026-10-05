"use client"

import { useCallback, useEffect, useState } from "react"

function decodificarValorArmazenado<T>(bruto: string): T | undefined {
  try {
    return JSON.parse(bruto) as T
  } catch {
    return undefined
  }
}

export function useLocalStorage<T>(chave: string, valorInicial: T) {
  const [valor, setValorState] = useState<T>(valorInicial)
  const [carregado, setCarregado] = useState(false)

  useEffect(() => {
    if (typeof window === "undefined") return
    const bruto = window.localStorage.getItem(chave)
    if (bruto !== null) {
      const decodificado = decodificarValorArmazenado<T>(bruto)
      if (decodificado !== undefined) {
        setValorState(decodificado)
      }
    }
    setCarregado(true)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [chave])

  const setValor = useCallback(
    (novo: T) => {
      setValorState(novo)
      if (typeof window !== "undefined") {
        window.localStorage.setItem(chave, JSON.stringify(novo))
      }
    },
    [chave]
  )

  const remover = useCallback(() => {
    setValorState(valorInicial)
    if (typeof window !== "undefined") {
      window.localStorage.removeItem(chave)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [chave])

  return { valor, setValor, remover, carregado }
}
