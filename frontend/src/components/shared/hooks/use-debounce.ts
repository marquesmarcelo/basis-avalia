"use client"

import { useEffect, useState } from "react"

export function useDebounce<T>(valor: T, atrasoMs: number): T {
  const [valorComAtraso, setValorComAtraso] = useState(valor)

  useEffect(() => {
    const temporizador = setTimeout(() => setValorComAtraso(valor), atrasoMs)
    return () => clearTimeout(temporizador)
  }, [valor, atrasoMs])

  return valorComAtraso
}
