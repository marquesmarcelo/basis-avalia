"use client"

import { useEffect } from "react"

const CAMPO_DE_BUSCA_POR_ROTA: Record<string, string> = {
  "/app/usuarios": "usuarios-filtro-busca",
  "/app/instituicoes": "instituicoes-filtro-busca",
  "/app/administradores": "administradores-filtro-busca",
  "/app/indicadores-inep": "indicadores-inep-filtro-busca",
  "/app/indicadores": "indicadores-filtro-busca",
  "/app/metas": "metas-filtro-busca",
}

export function useAtalhosDeTeclado() {
  useEffect(() => {
    function handler(e: KeyboardEvent) {
      const mod = e.ctrlKey || e.metaKey
      if (!mod) return

      if (e.key === "n") {
        const botao = document.getElementById("botao-novo")
        if (botao) {
          e.preventDefault()
          botao.click()
        }
        return
      }

      if (e.key === "f") {
        const idCampo = CAMPO_DE_BUSCA_POR_ROTA[window.location.pathname]
        const campo = idCampo ? document.getElementById(idCampo) : null
        if (campo) {
          e.preventDefault()
          campo.focus()
        }
      }
    }

    document.addEventListener("keydown", handler)
    return () => document.removeEventListener("keydown", handler)
  }, [])
}
