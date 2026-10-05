"use client"

import { useState } from "react"
import { SelectComRotulo } from "@/components/shared/forms/select-com-rotulo"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { Field, FieldContent, FieldLabel } from "@/components/ui/field"
import type { FiltroRelatorio } from "../types"

const rotuloSituacao: Record<string, string> = {
  cumprida: "Cumprida",
  sem_responsavel: "Sem responsável",
  em_andamento: "Em andamento",
  em_correcao: "Em correção",
  nao_cumprida: "Não cumprida",
}

interface RelatorioFiltroProps {
  filtrosIniciais: FiltroRelatorio
  opcoesPeriodo: { id: string; nome: string }[]
  onPesquisar: (filtros: FiltroRelatorio) => void
  pesquisando: boolean
}

export function RelatorioFiltro({
  filtrosIniciais,
  opcoesPeriodo,
  onPesquisar,
  pesquisando,
}: RelatorioFiltroProps) {
  const [periodoId, setPeriodoId] = useState(filtrosIniciais.periodo_id)
  const [situacao, setSituacao] = useState(filtrosIniciais.situacao)

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    onPesquisar({ ...filtrosIniciais, periodo_id: periodoId, situacao })
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col gap-3 sm:flex-row sm:items-end sm:flex-wrap"
    >
      <Field className="sm:w-64">
        <FieldLabel htmlFor="desempenho-filtro-periodo">Período</FieldLabel>
        <FieldContent>
          <SelectComRotulo
            id="desempenho-filtro-periodo"
            value={periodoId}
            onValueChange={setPeriodoId}
            itens={opcoesPeriodo.map((p) => ({ value: p.id, label: p.nome }))}
            placeholder="Selecione o período"
          />
        </FieldContent>
      </Field>
      <Field className="sm:w-56">
        <FieldLabel htmlFor="desempenho-filtro-situacao">Situação</FieldLabel>
        <FieldContent>
          <SelectComRotulo
            id="desempenho-filtro-situacao"
            value={situacao || "todas"}
            onValueChange={(v) => setSituacao(v === "todas" ? "" : v)}
            itens={[
              { value: "todas", label: "Todas" },
              ...Object.entries(rotuloSituacao).map(([valor, rotulo]) => ({
                value: valor,
                label: rotulo,
              })),
            ]}
          />
        </FieldContent>
      </Field>
      <LoadingButton
        type="submit"
        loading={pesquisando}
        loadingText="Pesquisando..."
        disabled={!periodoId}
      >
        Pesquisar
      </LoadingButton>
    </form>
  )
}
