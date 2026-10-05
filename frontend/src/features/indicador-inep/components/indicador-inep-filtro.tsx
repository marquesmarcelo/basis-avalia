"use client"

import { useState } from "react"
import { SelectComRotulo } from "@/components/shared/forms/select-com-rotulo"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { Field, FieldContent, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import type { FiltroIndicadoresInep } from "../types"

interface IndicadorInepFiltroProps {
  filtrosIniciais: FiltroIndicadoresInep
  onPesquisar: (filtros: FiltroIndicadoresInep) => void
  pesquisando: boolean
}

export function IndicadorInepFiltro({
  filtrosIniciais,
  onPesquisar,
  pesquisando,
}: IndicadorInepFiltroProps) {
  const [busca, setBusca] = useState(filtrosIniciais.busca)
  const [situacao, setSituacao] = useState(filtrosIniciais.situacao)

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    onPesquisar({ busca, situacao })
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col gap-3 sm:flex-row sm:items-end sm:flex-wrap"
    >
      <Field className="sm:w-64">
        <FieldLabel htmlFor="indicadores-inep-filtro-busca">Código ou nome</FieldLabel>
        <FieldContent>
          <Input
            id="indicadores-inep-filtro-busca"
            value={busca}
            onChange={(e) => setBusca(e.target.value)}
          />
        </FieldContent>
      </Field>
      <Field className="sm:w-48">
        <FieldLabel htmlFor="indicadores-inep-filtro-situacao">Situação</FieldLabel>
        <FieldContent>
          <SelectComRotulo
            id="indicadores-inep-filtro-situacao"
            value={situacao}
            onValueChange={(v) => setSituacao(v as FiltroIndicadoresInep["situacao"])}
            itens={[
              { value: "todos", label: "Todos" },
              { value: "ativo", label: "Ativos" },
              { value: "inativo", label: "Inativos" },
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
