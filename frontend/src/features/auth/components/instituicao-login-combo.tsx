"use client"

import { useMemo } from "react"
import {
  ComboboxEntidade,
  type ItemComboboxEntidade,
} from "@/components/shared/forms/combobox-entidade"
import { useInstituicoesPublicas } from "../hooks/use-instituicoes-publicas"

export const VALOR_ADMINISTRACAO_SISTEMA = "administracao-sistema"

interface InstituicaoLoginComboProps {
  id?: string
  value: string | null
  onValueChange: (value: string | null) => void
  invalido?: boolean
  descricaoId?: string
}

export function InstituicaoLoginCombo({
  id,
  value,
  onValueChange,
  invalido,
  descricaoId,
}: InstituicaoLoginComboProps) {
  const { data, isLoading, error, recarregar } = useInstituicoesPublicas()

  const itens: ItemComboboxEntidade[] = useMemo(
    () => data.map((i) => ({ value: i.id, label: `${i.nome} (${i.sigla})` })),
    [data]
  )

  return (
    <ComboboxEntidade
      id={id}
      itens={itens}
      itemFixo={{ value: VALOR_ADMINISTRACAO_SISTEMA, label: "Administração do sistema" }}
      rotuloGrupo="Instituições"
      value={value}
      onValueChange={onValueChange}
      carregando={isLoading}
      erro={!!error}
      onTentarNovamente={recarregar}
      placeholder="Selecione sua instituição"
      mensagemVazia="Nenhuma instituição cadastrada ainda."
      mensagemErro="Não foi possível carregar as instituições."
      invalido={invalido}
      descricaoId={descricaoId}
    />
  )
}
