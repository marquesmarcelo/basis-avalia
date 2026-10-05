"use client"

import { useEffect, useId, useRef, useState } from "react"
import { ComboboxEntidade } from "@/components/shared/forms/combobox-entidade"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Field, FieldContent, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { useMetasSugestoes } from "@/features/meta/hooks/use-metas-sugestoes"
import { useAtualizarItem } from "../hooks/use-atualizar-item"
import { useCriarItem } from "../hooks/use-criar-item"
import type { ItemDoPlano, Plano } from "../types"

interface ItemPlanoFormModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  planoId: string
  item: ItemDoPlano | null
  onSucesso: (item: Plano) => void
}

export function ItemPlanoFormModal({
  open,
  onOpenChange,
  planoId,
  item,
  onSucesso,
}: ItemPlanoFormModalProps) {
  const modoEdicao = !!item
  const idBase = useId()
  const primeiroCampoRef = useRef<HTMLButtonElement>(null)

  const [metaId, setMetaId] = useState<string | null>(item?.meta_id ?? null)
  const [quantidade, setQuantidade] = useState(item ? String(item.quantidade) : "1")
  const [erroMeta, setErroMeta] = useState<string | undefined>()
  const [erroQuantidade, setErroQuantidade] = useState<string | undefined>()

  const { itens: metas, isLoading: carregandoMetas, buscar: buscarMetas } = useMetasSugestoes()
  const {
    criar,
    isSubmitting: criando,
    error: erroCriar,
    limparErro: limparErroCriar,
  } = useCriarItem()
  const {
    atualizar,
    isSubmitting: atualizando,
    error: erroAtualizar,
    limparErro: limparErroAtualizar,
  } = useAtualizarItem()

  const salvando = criando || atualizando
  const erro = erroCriar ?? erroAtualizar
  const metaSelecionada = metas.find((m) => m.id === metaId) ?? item

  useEffect(() => {
    if (open) {
      setMetaId(item?.meta_id ?? null)
      setQuantidade(item ? String(item.quantidade) : "1")
      setErroMeta(undefined)
      setErroQuantidade(undefined)
      limparErroCriar()
      limparErroAtualizar()
      if (!modoEdicao) buscarMetas("")
      requestAnimationFrame(() => primeiroCampoRef.current?.focus())
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, item])

  useEffect(() => {
    if (erro?.code === "META_DUPLICADA_NO_PLANO") setErroMeta("Esta meta já está neste plano.")
    if (erro?.code === "QUANTIDADE_INVALIDA")
      setErroQuantidade("Informe uma quantidade de 1 ou mais.")
    if (erro?.code === "META_INATIVA")
      setErroMeta("Meta inativa não pode ser incluída em novo item.")
  }, [erro])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const numero = Number(quantidade)
    if (!modoEdicao && !metaId) {
      setErroMeta("Selecione uma meta.")
      return
    }
    if (!Number.isInteger(numero) || numero < 1) {
      setErroQuantidade("Informe uma quantidade de 1 ou mais.")
      return
    }
    setErroMeta(undefined)
    setErroQuantidade(undefined)

    if (modoEdicao && item) {
      const resultado = await atualizar(planoId, item.id, numero, item.versao)
      if (resultado) {
        notificar.sucesso("Quantidade atualizada.")
        onSucesso(resultado)
      }
    } else if (metaId) {
      const resultado = await criar(planoId, { meta_id: metaId, quantidade: numero })
      if (resultado) {
        notificar.sucesso("Meta adicionada ao plano.")
        onSucesso(resultado)
      }
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>
            {modoEdicao ? "Editar item do plano" : "Adicionar meta ao plano"}
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} aria-busy={salvando}>
          <FieldGroup>
            <Field data-invalid={!!erroMeta}>
              <FieldLabel htmlFor={`${idBase}-meta`}>Meta</FieldLabel>
              <FieldContent>
                {modoEdicao ? (
                  <p className="text-sm">{item?.meta_nome}</p>
                ) : (
                  <ComboboxEntidade
                    id={`${idBase}-meta`}
                    itens={metas.map((m) => ({ value: m.id, label: m.nome }))}
                    value={metaId}
                    onValueChange={(v) => {
                      setMetaId(v)
                      const escolhida = metas.find((m) => m.id === v)
                      if (escolhida?.quantidade_sugerida != null) {
                        setQuantidade(String(escolhida.quantidade_sugerida))
                      }
                    }}
                    carregando={carregandoMetas}
                    onTentarNovamente={() => buscarMetas("")}
                    disabled={salvando}
                  />
                )}
                {erroMeta && <FieldError>{erroMeta}</FieldError>}
              </FieldContent>
            </Field>

            {metaSelecionada &&
              "indicadores" in metaSelecionada &&
              metaSelecionada.indicadores.length > 0 && (
                <Field>
                  <FieldLabel>Indicadores da meta</FieldLabel>
                  <FieldContent>
                    <ul className="text-sm text-muted-foreground">
                      {metaSelecionada.indicadores.map((ind) => (
                        <li key={ind.id}>
                          {ind.codigo} — {ind.nome} (
                          {ind.escopo === "plataforma" ? "Do INEP" : "Próprio"})
                        </li>
                      ))}
                    </ul>
                  </FieldContent>
                </Field>
              )}

            <Field data-invalid={!!erroQuantidade}>
              <FieldLabel htmlFor={`${idBase}-quantidade`}>Quantidade</FieldLabel>
              <FieldContent>
                <Input
                  id={`${idBase}-quantidade`}
                  type="number"
                  min={1}
                  step={1}
                  value={quantidade}
                  onChange={(e) => setQuantidade(e.target.value)}
                  disabled={salvando}
                  aria-describedby={`${idBase}-quantidade-nota`}
                />
                <p id={`${idBase}-quantidade-nota`} className="text-sm text-muted-foreground">
                  Uma entrega atende todos os indicadores desta meta — a quantidade não é
                  multiplicada por eles. A quantidade é deste plano; a mesma meta pode exigir outro
                  número em outro curso.
                </p>
                {erroQuantidade && <FieldError>{erroQuantidade}</FieldError>}
              </FieldContent>
            </Field>
          </FieldGroup>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={salvando}
            >
              Cancelar
            </Button>
            <LoadingButton type="submit" loading={salvando} loadingText="Salvando...">
              Salvar
            </LoadingButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
