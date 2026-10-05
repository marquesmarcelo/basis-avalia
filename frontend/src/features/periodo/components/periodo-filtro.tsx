"use client"

import { useState } from "react"
import { SelectComRotulo } from "@/components/shared/forms/select-com-rotulo"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { Field, FieldContent, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import type { FiltroPeriodos } from "../types"

interface PeriodoFiltroProps {
  filtrosIniciais: FiltroPeriodos
  onPesquisar: (filtros: FiltroPeriodos) => void
  pesquisando: boolean
}

export function PeriodoFiltro({ filtrosIniciais, onPesquisar, pesquisando }: PeriodoFiltroProps) {
  const [nome, setNome] = useState(filtrosIniciais.nome)
  const [situacao, setSituacao] = useState(filtrosIniciais.situacao)

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    onPesquisar({ nome, situacao })
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col gap-3 sm:flex-row sm:items-end sm:flex-wrap"
    >
      <Field className="sm:w-64">
        <FieldLabel htmlFor="periodos-filtro-nome">Nome</FieldLabel>
        <FieldContent>
          <Input id="periodos-filtro-nome" value={nome} onChange={(e) => setNome(e.target.value)} />
        </FieldContent>
      </Field>
      <Field className="sm:w-48">
        <FieldLabel htmlFor="periodos-filtro-situacao">Situação</FieldLabel>
        <FieldContent>
          <SelectComRotulo
            id="periodos-filtro-situacao"
            value={situacao}
            onValueChange={(v) => setSituacao(v as FiltroPeriodos["situacao"])}
            itens={[
              { value: "todas", label: "Todas" },
              { value: "nao_iniciado", label: "Não iniciado" },
              { value: "aberto", label: "Aberto" },
              { value: "encerrado", label: "Encerrado" },
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
