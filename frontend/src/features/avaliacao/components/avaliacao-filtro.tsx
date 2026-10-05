"use client"

import { useState } from "react"
import { SelectComRotulo } from "@/components/shared/forms/select-com-rotulo"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { Field, FieldContent, FieldLabel } from "@/components/ui/field"
import type { FiltroFila } from "../hooks/use-fila-de-avaliacao"

interface AvaliacaoFiltroProps {
  filtrosIniciais: FiltroFila
  onPesquisar: (filtros: FiltroFila) => void
  pesquisando: boolean
}

export function AvaliacaoFiltro({
  filtrosIniciais,
  onPesquisar,
  pesquisando,
}: AvaliacaoFiltroProps) {
  const [situacao, setSituacao] = useState(filtrosIniciais.situacao)

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    onPesquisar({ ...filtrosIniciais, situacao })
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col gap-3 sm:flex-row sm:items-end sm:flex-wrap"
    >
      <Field className="sm:w-56">
        <FieldLabel htmlFor="avaliacoes-filtro-situacao">Situação</FieldLabel>
        <FieldContent>
          <SelectComRotulo
            id="avaliacoes-filtro-situacao"
            value={situacao || "todas"}
            onValueChange={(v) => setSituacao(v === "todas" ? "" : v)}
            itens={[
              { value: "pendente_avaliacao", label: "Pendente de avaliação" },
              { value: "aceita", label: "Aceita" },
              { value: "recusada", label: "Em correção" },
              { value: "todas", label: "Todas" },
            ]}
          />
        </FieldContent>
      </Field>
      <LoadingButton type="submit" loading={pesquisando} loadingText="Pesquisando...">
        Pesquisar
      </LoadingButton>
    </form>
  )
}
