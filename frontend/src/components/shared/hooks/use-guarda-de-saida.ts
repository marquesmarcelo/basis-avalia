"use client"

import { useRouter } from "next/navigation"
import { useEffect, useRef, useState } from "react"

export function useGuardaDeSaida(sujo: boolean) {
  const router = useRouter()
  const [confirmando, setConfirmando] = useState(false)
  const destinoRef = useRef<string | null>(null)

  useEffect(() => {
    if (!sujo) return

    function interceptarClique(e: MouseEvent) {
      const alvo = e.target as HTMLElement | null
      const ancora = alvo?.closest("a")
      if (!ancora) return
      if (ancora.target === "_blank") return
      const href = ancora.getAttribute("href")
      if (!href || href.startsWith("#")) return
      e.preventDefault()
      e.stopPropagation()
      destinoRef.current = href
      setConfirmando(true)
    }

    function interceptarVoltar() {
      window.history.pushState(null, "", window.location.href)
      destinoRef.current = "__voltar__"
      setConfirmando(true)
    }

    document.addEventListener("click", interceptarClique, true)
    window.history.pushState(null, "", window.location.href)
    window.addEventListener("popstate", interceptarVoltar)

    return () => {
      document.removeEventListener("click", interceptarClique, true)
      window.removeEventListener("popstate", interceptarVoltar)
    }
  }, [sujo])

  function confirmarDescarte() {
    setConfirmando(false)
    const destino = destinoRef.current
    destinoRef.current = null
    if (destino === "__voltar__") {
      window.history.back()
    } else if (destino) {
      router.push(destino)
    }
  }

  function cancelarDescarte() {
    setConfirmando(false)
    destinoRef.current = null
  }

  function solicitarSaida(destino: string) {
    if (!sujo) {
      router.push(destino)
      return
    }
    destinoRef.current = destino
    setConfirmando(true)
  }

  return { confirmando, confirmarDescarte, cancelarDescarte, solicitarSaida }
}
