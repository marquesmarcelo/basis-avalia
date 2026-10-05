"use client"

import { useEffect, useState } from "react"
import { ComboboxEntidade } from "@/components/shared/forms/combobox-entidade"
import { SelectComRotulo } from "@/components/shared/forms/select-com-rotulo"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { Field, FieldContent, FieldLabel } from "@/components/ui/field"
import { useCursosSugestoes } from "../hooks/use-cursos-sugestoes"
import { usePeriodosSugestoes } from "../hooks/use-periodos-sugestoes"
import type { FiltroPlanos } from "../types"

interface PlanoFiltroProps {
  filtrosIniciais: FiltroPlanos
  onPesquisar: (filtros: FiltroPlanos) => void
  pesquisando: boolean
  somenteLeitura?: boolean
}

export function PlanoFiltro({
  filtrosIniciais,
  onPesquisar,
  pesquisando,
  somenteLeitura,
}: PlanoFiltroProps) {
  const [periodoId, setPeriodoId] = useState<string | null>(filtrosIniciais.periodo_id || null)
  const [cursoId, setCursoId] = useState<string | null>(filtrosIniciais.curso_id || null)
  const [situacao, setSituacao] = useState(filtrosIniciais.situacao)
  const [aprovacao, setAprovacao] = useState(filtrosIniciais.aprovacao)

  const {
    itens: periodos,
    isLoading: carregandoPeriodos,
    buscar: buscarPeriodos,
  } = usePeriodosSugestoes()
  const { itens: cursos, isLoading: carregandoCursos, buscar: buscarCursos } = useCursosSugestoes()

  useEffect(() => {
    buscarPeriodos("")
    buscarCursos("")
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    onPesquisar({ periodo_id: periodoId ?? "", curso_id: cursoId ?? "", situacao, aprovacao })
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col gap-3 sm:flex-row sm:items-end sm:flex-wrap"
    >
      <Field className="sm:w-64">
        <FieldLabel htmlFor="planos-filtro-periodo">Período</FieldLabel>
        <FieldContent>
          <ComboboxEntidade
            id="planos-filtro-periodo"
            itens={periodos.map((p) => ({ value: p.id, label: p.nome }))}
            value={periodoId}
            onValueChange={setPeriodoId}
            carregando={carregandoPeriodos}
            onTentarNovamente={() => buscarPeriodos("")}
            placeholder="Todos"
          />
        </FieldContent>
      </Field>
      <Field className="sm:w-56">
        <FieldLabel htmlFor="planos-filtro-curso">Curso</FieldLabel>
        <FieldContent>
          <SelectComRotulo
            id="planos-filtro-curso"
            value={cursoId ?? "todos"}
            onValueChange={(v) => setCursoId(v === "todos" ? null : v)}
            disabled={carregandoCursos}
            itens={[
              { value: "todos", label: "Todos" },
              ...cursos.map((c) => ({ value: c.id, label: c.nome })),
            ]}
          />
        </FieldContent>
      </Field>
      <Field className="sm:w-44">
        <FieldLabel htmlFor="planos-filtro-situacao">Situação</FieldLabel>
        <FieldContent>
          <SelectComRotulo
            id="planos-filtro-situacao"
            value={situacao}
            onValueChange={(v) => setSituacao(v as FiltroPlanos["situacao"])}
            itens={[
              { value: "todas", label: "Todas" },
              { value: "rascunho", label: "Rascunho" },
              { value: "vigente", label: "Vigente" },
              { value: "encerrado", label: "Encerrado" },
            ]}
          />
        </FieldContent>
      </Field>
      {!somenteLeitura && (
        <Field className="sm:w-44">
          <FieldLabel htmlFor="planos-filtro-aprovacao">Aprovação</FieldLabel>
          <FieldContent>
            <SelectComRotulo
              id="planos-filtro-aprovacao"
              value={aprovacao}
              onValueChange={(v) => setAprovacao(v as FiltroPlanos["aprovacao"])}
              itens={[
                { value: "todos", label: "Todos" },
                { value: "aprovados", label: "Aprovados" },
                { value: "sem_aprovacao", label: "Sem aprovação" },
              ]}
            />
          </FieldContent>
        </Field>
      )}
      <LoadingButton type="submit" loading={pesquisando} loadingText="Pesquisando...">
        Pesquisar
      </LoadingButton>
    </form>
  )
}
