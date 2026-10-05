"use client"

import { useCallback, useEffect, useState } from "react"

export function useDirtyState<T extends Record<string, unknown>>(valoresIniciais: T) {
  const [valores, setValores] = useState<T>(valoresIniciais)
  const [sujo, setSujo] = useState(false)

  const definirCampo = useCallback(
    (chave: keyof T, valor: unknown) => {
      setValores((atual) => {
        const proximo = { ...atual, [chave]: valor }
        const ficouSujo = (Object.keys(proximo) as (keyof T)[]).some(
          (k) => proximo[k] !== valoresIniciais[k]
        )
        setSujo(ficouSujo)
        return proximo
      })
    },
    [valoresIniciais]
  )

  const reiniciar = useCallback(() => {
    setValores(valoresIniciais)
    setSujo(false)
  }, [valoresIniciais])

  useEffect(() => {
    if (!sujo) return
    const handler = (e: BeforeUnloadEvent) => {
      e.preventDefault()
    }
    window.addEventListener("beforeunload", handler)
    return () => window.removeEventListener("beforeunload", handler)
  }, [sujo])

  return { valores, definirCampo, sujo, reiniciar }
}
