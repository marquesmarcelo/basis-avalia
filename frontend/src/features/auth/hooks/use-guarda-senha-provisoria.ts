"use client"

import { useEffect } from "react"

export function useGuardaSenhaProvisoria(ativo: boolean) {
  useEffect(() => {
    if (!ativo) return

    function bloquearCliqueEmLink(e: MouseEvent) {
      const alvo = e.target as HTMLElement | null
      const ancora = alvo?.closest("a")
      if (ancora) {
        e.preventDefault()
        e.stopPropagation()
      }
    }

    function bloquearVoltar() {
      window.history.pushState(null, "", window.location.href)
    }

    document.addEventListener("click", bloquearCliqueEmLink, true)
    window.history.pushState(null, "", window.location.href)
    window.addEventListener("popstate", bloquearVoltar)

    return () => {
      document.removeEventListener("click", bloquearCliqueEmLink, true)
      window.removeEventListener("popstate", bloquearVoltar)
    }
  }, [ativo])
}
