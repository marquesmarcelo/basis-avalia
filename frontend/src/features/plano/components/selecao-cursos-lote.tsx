"use client"

import { useId } from "react"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import type { DestinoDeCopia } from "../types"

const LIMITE_LOTE = 100

interface SelecaoCursosLoteProps {
  destinos: DestinoDeCopia[]
  carregando: boolean
  busca: string
  onBuscaChange: (busca: string) => void
  selecionados: string[]
  onSelecionadosChange: (ids: string[]) => void
}

export function SelecaoCursosLote({
  destinos,
  carregando,
  busca,
  onBuscaChange,
  selecionados,
  onSelecionadosChange,
}: SelecaoCursosLoteProps) {
  const idBusca = useId()
  const selecionaveis = destinos.filter((d) => !d.ja_tem_plano)
  const todosSelecionados =
    selecionaveis.length > 0 && selecionaveis.every((d) => selecionados.includes(d.curso_id))
  const algumSelecionado = selecionados.length > 0 && !todosSelecionados

  function alternar(cursoId: string, marcado: boolean) {
    if (marcado) {
      if (selecionados.length >= LIMITE_LOTE) return
      onSelecionadosChange([...selecionados, cursoId])
    } else {
      onSelecionadosChange(selecionados.filter((id) => id !== cursoId))
    }
  }

  function alternarTodos(marcado: boolean) {
    if (!marcado) {
      onSelecionadosChange([])
      return
    }
    onSelecionadosChange(selecionaveis.slice(0, LIMITE_LOTE).map((d) => d.curso_id))
  }

  return (
    <div className="space-y-3">
      <div className="max-w-sm">
        <label htmlFor={idBusca} className="sr-only">
          Buscar curso
        </label>
        <Input
          id={idBusca}
          placeholder="Buscar curso..."
          value={busca}
          onChange={(e) => onBuscaChange(e.target.value)}
        />
      </div>

      <p aria-live="polite" className="text-sm font-medium">
        {selecionados.length} de {LIMITE_LOTE} selecionados
      </p>

      {carregando && (
        <div className="space-y-2">
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-10 w-full" />
        </div>
      )}

      {!carregando && destinos.length === 0 && (
        <p className="text-sm text-muted-foreground">
          Nenhum curso ativo encontrado para {busca || "a instituição"}.
        </p>
      )}

      {!carregando && destinos.length > 0 && (
        <div className="rounded-md border">
          <div className="flex items-center gap-2 border-b p-2">
            <Checkbox
              checked={todosSelecionados}
              indeterminate={algumSelecionado}
              onCheckedChange={(marcado) => alternarTodos(!!marcado)}
              aria-label={`Selecionar todos os selecionáveis (${selecionaveis.length})`}
            />
            <span className="text-sm">
              Selecionar todos os selecionáveis ({selecionaveis.length})
            </span>
          </div>
          <ul className="divide-y">
            {destinos.map((destino) => {
              const desabilitado =
                destino.ja_tem_plano ||
                (selecionados.length >= LIMITE_LOTE && !selecionados.includes(destino.curso_id))
              const idMotivo = `motivo-${destino.curso_id}`
              return (
                <li key={destino.curso_id} className="flex items-center gap-2 p-2">
                  <Checkbox
                    checked={selecionados.includes(destino.curso_id)}
                    onCheckedChange={(marcado) => alternar(destino.curso_id, !!marcado)}
                    disabled={desabilitado}
                    aria-disabled={desabilitado}
                    aria-describedby={desabilitado ? idMotivo : undefined}
                  />
                  <span className="flex-1 text-sm">{destino.curso_nome}</span>
                  <span className="text-sm text-muted-foreground">
                    {destino.vago ? "Vago ⚠" : (destino.coordenador_nome ?? "—")}
                  </span>
                  {(destino.ja_tem_plano || (desabilitado && !destino.ja_tem_plano)) && (
                    <span id={idMotivo} className="text-sm text-muted-foreground">
                      {destino.ja_tem_plano
                        ? "Já tem plano"
                        : "Limite de 100 cursos por lote atingido"}
                    </span>
                  )}
                </li>
              )
            })}
          </ul>
        </div>
      )}
    </div>
  )
}
