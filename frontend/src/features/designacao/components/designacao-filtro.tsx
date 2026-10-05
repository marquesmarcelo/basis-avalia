"use client"

import { useEffect, useState } from "react"
import { ComboboxEntidade } from "@/components/shared/forms/combobox-entidade"
import { SelectComRotulo } from "@/components/shared/forms/select-com-rotulo"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { Field, FieldContent, FieldLabel } from "@/components/ui/field"
import { useCandidatosDesignacao } from "../hooks/use-candidatos-designacao"
import type { FiltroDesignacoes } from "../types"

interface DesignacaoFiltroProps {
  filtrosIniciais: FiltroDesignacoes
  onPesquisar: (filtros: FiltroDesignacoes) => void
  pesquisando: boolean
}

export function DesignacaoFiltro({
  filtrosIniciais,
  onPesquisar,
  pesquisando,
}: DesignacaoFiltroProps) {
  const [situacao, setSituacao] = useState(filtrosIniciais.situacao)
  const [coordenadorId, setCoordenadorId] = useState<string | null>(
    filtrosIniciais.coordenador_id || null
  )

  const { itens, isLoading, error, buscar } = useCandidatosDesignacao()

  useEffect(() => {
    buscar("")
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    onPesquisar({ situacao, coordenador_id: coordenadorId ?? "" })
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col gap-3 sm:flex-row sm:items-end sm:flex-wrap"
    >
      <Field className="sm:w-44">
        <FieldLabel htmlFor="designacoes-filtro-situacao">Situação</FieldLabel>
        <FieldContent>
          <SelectComRotulo
            id="designacoes-filtro-situacao"
            value={situacao}
            onValueChange={(v) => setSituacao(v as FiltroDesignacoes["situacao"])}
            itens={[
              { value: "todas", label: "Todas" },
              { value: "futura", label: "Futura" },
              { value: "vigente", label: "Vigente" },
              { value: "encerrada", label: "Encerrada" },
            ]}
          />
        </FieldContent>
      </Field>
      <Field className="sm:w-64">
        <FieldLabel htmlFor="designacoes-filtro-coordenador">Coordenador</FieldLabel>
        <FieldContent>
          <ComboboxEntidade
            id="designacoes-filtro-coordenador"
            itens={itens.map((c) => ({ value: c.id, label: c.nome }))}
            value={coordenadorId}
            onValueChange={setCoordenadorId}
            carregando={isLoading}
            erro={!!error}
            onTentarNovamente={() => buscar("")}
            placeholder="Todos"
          />
        </FieldContent>
      </Field>
      <LoadingButton type="submit" loading={pesquisando} loadingText="Pesquisando...">
        Pesquisar
      </LoadingButton>
    </form>
  )
}
