"use client"

import { useState } from "react"
import { SelectComRotulo } from "@/components/shared/forms/select-com-rotulo"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { Field, FieldContent, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import type { FiltroIndicadores } from "../types"

interface IndicadorFiltroProps {
  filtrosIniciais: FiltroIndicadores
  onPesquisar: (filtros: FiltroIndicadores) => void
  pesquisando: boolean
}

export function IndicadorFiltro({
  filtrosIniciais,
  onPesquisar,
  pesquisando,
}: IndicadorFiltroProps) {
  const [busca, setBusca] = useState(filtrosIniciais.busca)
  const [origem, setOrigem] = useState(filtrosIniciais.origem)
  const [situacao, setSituacao] = useState(filtrosIniciais.situacao)

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    onPesquisar({ busca, origem, situacao })
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col gap-3 sm:flex-row sm:items-end sm:flex-wrap"
    >
      <Field className="sm:w-64">
        <FieldLabel htmlFor="indicadores-filtro-busca">Código ou nome</FieldLabel>
        <FieldContent>
          <Input
            id="indicadores-filtro-busca"
            value={busca}
            onChange={(e) => setBusca(e.target.value)}
          />
        </FieldContent>
      </Field>
      <Field className="sm:w-48">
        <FieldLabel htmlFor="indicadores-filtro-origem">Origem</FieldLabel>
        <FieldContent>
          <SelectComRotulo
            id="indicadores-filtro-origem"
            value={origem}
            onValueChange={(v) => setOrigem(v as FiltroIndicadores["origem"])}
            itens={[
              { value: "todos", label: "Todos" },
              { value: "plataforma", label: "Do INEP" },
              { value: "instituicao", label: "Próprios" },
            ]}
          />
        </FieldContent>
      </Field>
      <Field className="sm:w-48">
        <FieldLabel htmlFor="indicadores-filtro-situacao">Situação</FieldLabel>
        <FieldContent>
          <SelectComRotulo
            id="indicadores-filtro-situacao"
            value={situacao}
            onValueChange={(v) => setSituacao(v as FiltroIndicadores["situacao"])}
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
