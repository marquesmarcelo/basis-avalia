"use client"

import { useEffect, useState } from "react"
import { ComboboxEntidade } from "@/components/shared/forms/combobox-entidade"
import { SelectComRotulo } from "@/components/shared/forms/select-com-rotulo"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { Field, FieldContent, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { useCandidatosDesignacao } from "@/features/designacao/hooks/use-candidatos-designacao"
import type { FiltroCursos } from "../types"

interface CursoFiltroProps {
  filtrosIniciais: FiltroCursos
  onPesquisar: (filtros: FiltroCursos) => void
  pesquisando: boolean
}

export function CursoFiltro({ filtrosIniciais, onPesquisar, pesquisando }: CursoFiltroProps) {
  const [busca, setBusca] = useState(filtrosIniciais.busca)
  const [grau, setGrau] = useState(filtrosIniciais.grau)
  const [modalidade, setModalidade] = useState(filtrosIniciais.modalidade)
  const [coordenadorId, setCoordenadorId] = useState<string | null>(
    filtrosIniciais.coordenador_id || null
  )
  const [situacao, setSituacao] = useState(filtrosIniciais.situacao)

  const { itens, isLoading, error, buscar } = useCandidatosDesignacao()

  useEffect(() => {
    buscar("")
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    onPesquisar({ busca, grau, modalidade, coordenador_id: coordenadorId ?? "", situacao })
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col gap-3 sm:flex-row sm:items-end sm:flex-wrap"
    >
      <Field className="sm:w-64">
        <FieldLabel htmlFor="cursos-filtro-busca">Nome ou código</FieldLabel>
        <FieldContent>
          <Input
            id="cursos-filtro-busca"
            value={busca}
            onChange={(e) => setBusca(e.target.value)}
          />
        </FieldContent>
      </Field>
      <Field className="sm:w-44">
        <FieldLabel htmlFor="cursos-filtro-grau">Grau</FieldLabel>
        <FieldContent>
          <SelectComRotulo
            id="cursos-filtro-grau"
            value={grau}
            onValueChange={(v) => setGrau(v as FiltroCursos["grau"])}
            itens={[
              { value: "todos", label: "Todos" },
              { value: "bacharelado", label: "Bacharelado" },
              { value: "licenciatura", label: "Licenciatura" },
              { value: "tecnologo", label: "Tecnólogo" },
            ]}
          />
        </FieldContent>
      </Field>
      <Field className="sm:w-44">
        <FieldLabel htmlFor="cursos-filtro-modalidade">Modalidade</FieldLabel>
        <FieldContent>
          <SelectComRotulo
            id="cursos-filtro-modalidade"
            value={modalidade}
            onValueChange={(v) => setModalidade(v as FiltroCursos["modalidade"])}
            itens={[
              { value: "todas", label: "Todas" },
              { value: "presencial", label: "Presencial" },
              { value: "a_distancia", label: "A distância" },
            ]}
          />
        </FieldContent>
      </Field>
      <Field className="sm:w-64">
        <FieldLabel htmlFor="cursos-filtro-coordenador">Coordenador</FieldLabel>
        <FieldContent>
          <ComboboxEntidade
            id="cursos-filtro-coordenador"
            itens={itens.map((c) => ({ value: c.id, label: c.nome }))}
            itemFixo={{ value: "__vago__", label: "Vago" }}
            value={coordenadorId}
            onValueChange={setCoordenadorId}
            carregando={isLoading}
            erro={!!error}
            onTentarNovamente={() => buscar("")}
            placeholder="Todos"
          />
        </FieldContent>
      </Field>
      <Field className="sm:w-44">
        <FieldLabel htmlFor="cursos-filtro-situacao">Situação</FieldLabel>
        <FieldContent>
          <SelectComRotulo
            id="cursos-filtro-situacao"
            value={situacao}
            onValueChange={(v) => setSituacao(v as FiltroCursos["situacao"])}
            itens={[
              { value: "todas", label: "Todas" },
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
