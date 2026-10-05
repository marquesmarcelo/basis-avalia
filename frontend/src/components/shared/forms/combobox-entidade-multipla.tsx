"use client"

import { Combobox as ComboboxPrimitive } from "@base-ui/react"
import type { VariantProps } from "class-variance-authority"
import { XIcon } from "lucide-react"
import * as React from "react"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import type { badgeVariants } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Combobox,
  ComboboxContent,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
  ComboboxStatus,
} from "@/components/ui/combobox"
import { Skeleton } from "@/components/ui/skeleton"

export interface ItemComboboxMultiplo {
  value: string
  label: string
  badge?: string
  badgeVariant?: VariantProps<typeof badgeVariants>["variant"]
  secundario?: string
}

interface ComboboxEntidadeMultiplaProps {
  id?: string
  itens: ItemComboboxMultiplo[]
  value: string[]
  onValueChange: (value: string[]) => void
  max: number
  min?: number
  carregando?: boolean
  erro?: boolean
  onTentarNovamente?: () => void
  placeholder?: string
  mensagemVazia?: string
  mensagemSemItens?: React.ReactNode
  disabled?: boolean
}

function EstadoDaListaDinamica({
  carregando,
  erro,
  mensagemVazia,
}: {
  carregando: boolean
  erro: boolean
  mensagemVazia: string
}) {
  const filtrados = ComboboxPrimitive.useFilteredItems<ItemComboboxMultiplo>()
  if (carregando || erro || filtrados.length > 0) return null
  return <div className="px-2 py-3 text-center text-sm text-muted-foreground">{mensagemVazia}</div>
}

export function ComboboxEntidadeMultipla({
  id,
  itens,
  value,
  onValueChange,
  max,
  min = 1,
  carregando = false,
  erro = false,
  onTentarNovamente,
  placeholder = "Buscar...",
  mensagemVazia = "Nenhum resultado encontrado.",
  mensagemSemItens,
  disabled,
}: ComboboxEntidadeMultiplaProps) {
  const idBase = React.useId()
  const mapaPorId = React.useMemo(
    () => Object.fromEntries(itens.map((item) => [item.value, item])),
    [itens]
  )
  const { contains } = ComboboxPrimitive.useFilter()
  const filtrarItem = React.useCallback(
    (item: ItemComboboxMultiplo, query: string) => contains(item.label, query),
    [contains]
  )

  if (!carregando && !erro && itens.length === 0 && mensagemSemItens) {
    return <div>{mensagemSemItens}</div>
  }

  return (
    <div>
      <Combobox
        items={itens}
        multiple
        value={value}
        onValueChange={(v) => {
          const novos = v as string[]
          if (novos.length <= max) onValueChange(novos)
        }}
        itemToStringLabel={(v) => mapaPorId[v as string]?.label ?? ""}
        filter={filtrarItem}
      >
        <ComboboxInput
          id={id}
          placeholder={placeholder}
          showTrigger={false}
          showClear={false}
          disabled={disabled || carregando || erro || value.length >= max}
          aria-describedby={`${idBase}-contador ${idBase}-limite`}
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
                Não foi possível carregar as opções.
                {onTentarNovamente && (
                  <Button type="button" variant="outline" size="sm" onClick={onTentarNovamente}>
                    Tentar novamente
                  </Button>
                )}
              </div>
            )}
          </ComboboxStatus>
          <ComboboxList>
            {(item: ItemComboboxMultiplo) => (
              <ComboboxItem key={item.value} value={item.value}>
                <span>{item.label}</span>
                {item.badge && (
                  <StatusBadge label={item.badge} variant={item.badgeVariant} className="ml-auto" />
                )}
              </ComboboxItem>
            )}
          </ComboboxList>
          {!carregando && !erro && (
            <EstadoDaListaDinamica
              carregando={carregando}
              erro={erro}
              mensagemVazia={mensagemVazia}
            />
          )}
        </ComboboxContent>
      </Combobox>

      <ul aria-label="Indicadores selecionados" className="mt-2 divide-y rounded-md border">
        {value.map((id) => {
          const item = mapaPorId[id]
          if (!item) return null
          return (
            <li key={id} className="flex items-start justify-between gap-2 p-2">
              <div>
                <span className="text-sm">{item.label}</span>
                {item.badge && (
                  <StatusBadge label={item.badge} variant={item.badgeVariant} className="ml-2" />
                )}
                {item.secundario && (
                  <p className="text-xs text-muted-foreground">{item.secundario}</p>
                )}
              </div>
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                aria-label={`Remover ${item.label} da meta`}
                disabled={disabled}
                onClick={() => onValueChange(value.filter((v) => v !== id))}
              >
                <XIcon aria-hidden="true" />
              </Button>
            </li>
          )
        })}
      </ul>

      <p
        id={`${idBase}-contador`}
        aria-live="polite"
        className="mt-1 text-sm text-muted-foreground"
      >
        {value.length} de {max} indicadores selecionados
      </p>
      {value.length >= max && (
        <p id={`${idBase}-limite`} className="text-sm text-muted-foreground">
          Limite de {max} indicadores atingido. Remova um para adicionar outro.
        </p>
      )}
      {value.length < min && <p className="sr-only">Selecione pelo menos {min} indicador.</p>}
    </div>
  )
}
