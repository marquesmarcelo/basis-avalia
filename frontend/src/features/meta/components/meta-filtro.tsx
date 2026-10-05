"use client"

import { useEffect, useState } from "react"
import { ComboboxEntidade } from "@/components/shared/forms/combobox-entidade"
import { SelectComRotulo } from "@/components/shared/forms/select-com-rotulo"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { Field, FieldContent, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { useIndicadoresSugestoes } from "@/features/indicador/hooks/use-indicadores-sugestoes"
import type { FiltroMetas } from "../types"

interface MetaFiltroProps {
  filtrosIniciais: FiltroMetas
  onPesquisar: (filtros: FiltroMetas) => void
  pesquisando: boolean
}

export function MetaFiltro({ filtrosIniciais, onPesquisar, pesquisando }: MetaFiltroProps) {
  const [busca, setBusca] = useState(filtrosIniciais.busca)
  const [indicadorId, setIndicadorId] = useState<string | null>(
    filtrosIniciais.indicador_id || null
  )
  const [origem, setOrigem] = useState(filtrosIniciais.origem)
  const [situacao, setSituacao] = useState(filtrosIniciais.situacao)

  const {
    itens,
    isLoading: carregandoIndicadores,
    error: erroIndicadores,
    buscar,
  } = useIndicadoresSugestoes()

  useEffect(() => {
    buscar("")
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    onPesquisar({ busca, indicador_id: indicadorId ?? "", origem, situacao })
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col gap-3 sm:flex-row sm:items-end sm:flex-wrap"
    >
      <Field className="sm:w-64">
        <FieldLabel htmlFor="metas-filtro-busca">Nome</FieldLabel>
        <FieldContent>
          <Input id="metas-filtro-busca" value={busca} onChange={(e) => setBusca(e.target.value)} />
        </FieldContent>
      </Field>
      <Field className="sm:w-72">
        <FieldLabel htmlFor="metas-filtro-indicador">Indicador</FieldLabel>
        <FieldContent>
          <ComboboxEntidade
            id="metas-filtro-indicador"
            itens={itens.map((i) => ({
              value: i.id,
              label: `${i.codigo} — ${i.nome} (${i.escopo === "plataforma" ? "Do INEP" : "Próprio"})`,
            }))}
            value={indicadorId}
            onValueChange={setIndicadorId}
            carregando={carregandoIndicadores}
            erro={!!erroIndicadores}
            onTentarNovamente={() => buscar("")}
            placeholder="Todos"
          />
        </FieldContent>
      </Field>
      <Field className="sm:w-48">
        <FieldLabel htmlFor="metas-filtro-origem">Origem</FieldLabel>
        <FieldContent>
          <SelectComRotulo
            id="metas-filtro-origem"
            value={origem}
            onValueChange={(v) => setOrigem(v as FiltroMetas["origem"])}
            itens={[
              { value: "todas", label: "Todas" },
              { value: "plataforma", label: "Do INEP" },
              { value: "instituicao", label: "Próprios" },
            ]}
          />
        </FieldContent>
      </Field>
      <Field className="sm:w-48">
        <FieldLabel htmlFor="metas-filtro-situacao">Situação</FieldLabel>
        <FieldContent>
          <SelectComRotulo
            id="metas-filtro-situacao"
            value={situacao}
            onValueChange={(v) => setSituacao(v as FiltroMetas["situacao"])}
            itens={[
              { value: "todas", label: "Todas" },
              { value: "ativo", label: "Ativas" },
              { value: "inativo", label: "Inativas" },
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
