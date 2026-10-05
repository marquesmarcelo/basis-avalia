"use client"

import { useState } from "react"
import type { ParametrosListagem } from "@/lib/tipos-api"
import type { FiltroRelatorio } from "../types"
import { useRelatorioDesempenho } from "./use-relatorio-desempenho"

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:3001/api/v1"

export function useExportarRelatorio() {
  const [isExporting, setIsExporting] = useState(false)
  const { montarQuery } = useRelatorioDesempenho()

  async function exportar(
    filtros: FiltroRelatorio,
    paginacao: ParametrosListagem
  ): Promise<boolean> {
    if (!filtros.periodo_id) return false
    setIsExporting(true)
    try {
      const query = montarQuery(filtros, paginacao)
      const resposta = await fetch(
        `${BASE_URL}/relatorios/desempenho/exportacao?${query.toString()}`,
        {
          credentials: "include",
        }
      )
      if (!resposta.ok) return false
      const blob = await resposta.blob()
      const url = URL.createObjectURL(blob)
      const link = document.createElement("a")
      link.href = url
      link.download = "relatorio-desempenho.csv"
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      URL.revokeObjectURL(url)
      return true
    } catch {
      return false
    } finally {
      setIsExporting(false)
    }
  }

  return { exportar, isExporting }
}
