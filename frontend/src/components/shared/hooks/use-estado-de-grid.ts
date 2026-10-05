"use client"

import { useCallback } from "react"
import { useLocalStorage } from "./use-local-storage"

export interface EstadoDeGrid<TFiltro> {
  filtros: TFiltro
  sort: string
  order: "asc" | "desc"
  pageSize: number
  page: number
}

export function useEstadoDeGrid<TFiltro extends Record<string, unknown>>(
  entidade: string,
  filtrosIniciais: TFiltro,
  sortPadrao: string,
  orderPadrao: "asc" | "desc" = "asc",
  pageSizePadrao = 20
) {
  const chave = `grid-state:${entidade}`
  const { valor, setValor, carregado } = useLocalStorage<EstadoDeGrid<TFiltro>>(chave, {
    filtros: filtrosIniciais,
    sort: sortPadrao,
    order: orderPadrao,
    pageSize: pageSizePadrao,
    page: 1,
  })

  const definirFiltros = useCallback(
    (filtros: TFiltro) => setValor({ ...valor, filtros, page: 1 }),
    [valor, setValor]
  )

  const ordenarPor = useCallback(
    (campo: string) => {
      if (valor.sort === campo) {
        setValor({ ...valor, order: valor.order === "asc" ? "desc" : "asc" })
      } else {
        setValor({ ...valor, sort: campo, order: "asc" })
      }
    },
    [valor, setValor]
  )

  const definirPagina = useCallback(
    (page: number) => setValor({ ...valor, page }),
    [valor, setValor]
  )

  const definirTamanhoDePagina = useCallback(
    (pageSize: number) => setValor({ ...valor, pageSize, page: 1 }),
    [valor, setValor]
  )

  const ajustarPaginaAoTotal = useCallback(
    (totalPages: number) => {
      if (totalPages > 0 && valor.page > totalPages) {
        setValor({ ...valor, page: totalPages })
      }
    },
    [valor, setValor]
  )

  return {
    estado: valor,
    carregado,
    definirFiltros,
    ordenarPor,
    definirPagina,
    definirTamanhoDePagina,
    ajustarPaginaAoTotal,
  }
}
