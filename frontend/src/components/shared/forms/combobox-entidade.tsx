"use client"

import { Combobox as ComboboxPrimitive } from "@base-ui/react"
import type { VariantProps } from "class-variance-authority"
import * as React from "react"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import type { badgeVariants } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Combobox,
  ComboboxCollection,
  ComboboxContent,
  ComboboxGroup,
  ComboboxInput,
  ComboboxItem,
  ComboboxLabel,
  ComboboxList,
  ComboboxSeparator,
  ComboboxStatus,
} from "@/components/ui/combobox"
import { Skeleton } from "@/components/ui/skeleton"

export interface ItemComboboxEntidade {
  value: string
  label: string
  badge?: string
  badgeVariant?: VariantProps<typeof badgeVariants>["variant"]
}

interface ComboboxEntidadeProps {
  id?: string
  itens: ItemComboboxEntidade[]
  itemFixo?: ItemComboboxEntidade
  rotuloGrupo?: string
  value: string | null
  onValueChange: (value: string | null) => void
  carregando?: boolean
  erro?: boolean
  onTentarNovamente?: () => void
  placeholder?: string
  mensagemVazia?: string
  mensagemErro?: string
  invalido?: boolean
  descricaoId?: string
  disabled?: boolean
}

interface GrupoComboboxEntidade {
  chave: string
  rotulo?: string
  fixo: boolean
  itens: ItemComboboxEntidade[]
}

function EstadoDaListaDinamica({
  valorDoItemFixo,
  carregando,
  erro,
  mensagemVazia,
}: {
  valorDoItemFixo?: string
  carregando: boolean
  erro: boolean
  mensagemVazia: string
}) {
  const filtrados = ComboboxPrimitive.useFilteredItems<ItemComboboxEntidade>()
  if (carregando || erro) return null
  const temResultadoDinamico = filtrados.some((item) => item.value !== valorDoItemFixo)
  if (temResultadoDinamico) return null
  return <div className="px-2 py-3 text-center text-sm text-muted-foreground">{mensagemVazia}</div>
}

export function ComboboxEntidade({
  id,
  itens,
  itemFixo,
  rotuloGrupo,
  value,
  onValueChange,
  carregando = false,
  erro = false,
  onTentarNovamente,
  placeholder = "Selecione",
  mensagemVazia = "Nenhum resultado encontrado.",
  mensagemErro = "Não foi possível carregar as opções.",
  invalido,
  descricaoId,
  disabled,
}: ComboboxEntidadeProps) {
  const todosOsItens = React.useMemo(
    () => (itemFixo ? [itemFixo, ...itens] : itens),
    [itemFixo, itens]
  )

  const grupos = React.useMemo<GrupoComboboxEntidade[]>(() => {
    const lista: GrupoComboboxEntidade[] = []
    if (itemFixo) {
      lista.push({ chave: "fixo", fixo: true, itens: [itemFixo] })
    }
    lista.push({ chave: "dinamico", rotulo: rotuloGrupo, fixo: false, itens })
    return lista
  }, [itemFixo, rotuloGrupo, itens])

  const { contains } = ComboboxPrimitive.useFilter()

  const filtrarItem = React.useCallback(
    (item: ItemComboboxEntidade, query: string) => {
      if (itemFixo && item.value === itemFixo.value) return true
      return contains(item.label, query)
    },
    [contains, itemFixo]
  )

  return (
    <Combobox
      items={grupos}
      value={value}
      onValueChange={(v) => onValueChange((v as string | null) ?? null)}
      itemToStringLabel={(v) => todosOsItens.find((i) => i.value === v)?.label ?? ""}
      filter={filtrarItem}
    >
      <ComboboxInput
        id={id}
        placeholder={placeholder}
        showClear={!!value}
        disabled={disabled}
        aria-invalid={invalido}
        aria-describedby={descricaoId}
        aria-busy={carregando}
      />
      <ComboboxContent>
        <ComboboxStatus>
          {carregando && (
            <div className="flex flex-col gap-1 p-1" aria-hidden="true">
              <Skeleton className="h-8 w-full" />
              <Skeleton className="h-8 w-full" />
            </div>
          )}
          {!carregando && erro && (
            <div
              role="alert"
              className="flex flex-col items-center gap-2 p-3 text-center text-sm text-muted-foreground"
            >
              {mensagemErro}
              {onTentarNovamente && (
                <Button type="button" variant="outline" size="sm" onClick={onTentarNovamente}>
                  Tentar novamente
                </Button>
              )}
            </div>
          )}
        </ComboboxStatus>

        <ComboboxList>
          {(grupo: GrupoComboboxEntidade) => (
            <React.Fragment key={grupo.chave}>
              <ComboboxGroup items={grupo.itens}>
                {grupo.rotulo && !carregando && !erro && (
                  <ComboboxLabel>{grupo.rotulo}</ComboboxLabel>
                )}
                {(grupo.fixo || (!carregando && !erro)) && (
                  <ComboboxCollection>
                    {(item: ItemComboboxEntidade) => (
                      <ComboboxItem key={item.value} value={item.value}>
                        <span className="flex-1">{item.label}</span>
                        {item.badge && (
                          <StatusBadge
                            label={item.badge}
                            variant={item.badgeVariant}
                            className="ml-auto"
                          />
                        )}
                      </ComboboxItem>
                    )}
                  </ComboboxCollection>
                )}
              </ComboboxGroup>
              {grupo.fixo && <ComboboxSeparator />}
              {!grupo.fixo && (
                <EstadoDaListaDinamica
                  valorDoItemFixo={itemFixo?.value}
                  carregando={carregando}
                  erro={erro}
                  mensagemVazia={mensagemVazia}
                />
              )}
            </React.Fragment>
          )}
        </ComboboxList>
      </ComboboxContent>
    </Combobox>
  )
}
