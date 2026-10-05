"use client"

import { useState } from "react"
import { SelectComRotulo } from "@/components/shared/forms/select-com-rotulo"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { Field, FieldContent, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import type { FiltroInstituicoes } from "../types"

interface InstituicaoFiltroProps {
  filtrosIniciais: FiltroInstituicoes
  onPesquisar: (filtros: FiltroInstituicoes) => void
  pesquisando: boolean
}

export function InstituicaoFiltro({
  filtrosIniciais,
  onPesquisar,
  pesquisando,
}: InstituicaoFiltroProps) {
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
        <FieldLabel htmlFor="instituicoes-filtro-busca">Nome ou sigla</FieldLabel>
        <FieldContent>
          <Input
            id="instituicoes-filtro-busca"
            value={busca}
            onChange={(e) => setBusca(e.target.value)}
          />
        </FieldContent>
      </Field>
      <Field className="sm:w-48">
        <FieldLabel htmlFor="instituicoes-filtro-situacao">Situação</FieldLabel>
        <FieldContent>
          <SelectComRotulo
            id="instituicoes-filtro-situacao"
            value={situacao}
            onValueChange={(v) => setSituacao(v as FiltroInstituicoes["situacao"])}
            itens={[
              { value: "todas", label: "Todas" },
              { value: "ativa", label: "Ativa" },
              { value: "inativa", label: "Inativa" },
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
