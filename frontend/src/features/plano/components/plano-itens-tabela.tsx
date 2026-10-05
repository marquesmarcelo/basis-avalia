"use client"

import { PencilIcon, PlusIcon, Trash2Icon } from "lucide-react"
import { useState } from "react"
import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import { EmptyState } from "@/components/shared/ui/empty-state"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { useExcluirItem } from "../hooks/use-excluir-item"
import type { ItemDoPlano, Plano } from "../types"
import { ItemPlanoFormModal } from "./item-plano-form-modal"

interface PlanoItensTabelaProps {
  plano: Plano
  somenteLeitura: boolean
  onAtualizado: (item: Plano) => void
}

export function PlanoItensTabela({ plano, somenteLeitura, onAtualizado }: PlanoItensTabelaProps) {
  const [modalAberto, setModalAberto] = useState(false)
  const [emEdicao, setEmEdicao] = useState<ItemDoPlano | null>(null)
  const [paraExcluir, setParaExcluir] = useState<ItemDoPlano | null>(null)
  const { excluir, idEmAndamento: idExcluindo } = useExcluirItem()

  const itens = plano.itens ?? []
  const totalExigido = itens.reduce((soma, item) => soma + item.quantidade, 0)

  function abrirNovo() {
    setEmEdicao(null)
    setModalAberto(true)
  }

  function abrirEdicao(item: ItemDoPlano) {
    setEmEdicao(item)
    setModalAberto(true)
  }

  async function confirmarExclusao() {
    if (!paraExcluir) return
    const resultado = await excluir(plano.id, paraExcluir.id)
    if (resultado) {
      notificar.sucesso("Meta removida do plano.")
      setParaExcluir(null)
      onAtualizado(resultado)
    } else {
      notificar.erro(
        "Este item já tem entregas registradas — reduza a quantidade em vez de removê-lo."
      )
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-medium">Metas do plano</h2>
        {!somenteLeitura && (
          <Button size="sm" onClick={abrirNovo}>
            <PlusIcon aria-hidden="true" /> Adicionar meta
          </Button>
        )}
      </div>

      {itens.length === 0 && (
        <EmptyState
          titulo="Nenhuma meta adicionada ainda."
          descricao={
            somenteLeitura
              ? undefined
              : "Adicione pelo menos uma meta para poder publicar este plano."
          }
        />
      )}

      {itens.length > 0 && (
        <>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Meta</TableHead>
                <TableHead>Indicadores</TableHead>
                <TableHead>Qtd</TableHead>
                {!somenteLeitura && <TableHead>Ações</TableHead>}
              </TableRow>
            </TableHeader>
            <TableBody>
              {itens.map((item) => (
                <TableRow key={item.id}>
                  <TableCell>{item.meta_nome}</TableCell>
                  <TableCell>
                    <div className="flex flex-wrap gap-1 text-sm">
                      {item.indicadores.map((ind) => (
                        <span key={ind.id}>
                          {ind.codigo} ({ind.escopo === "plataforma" ? "Do INEP" : "Próprio"})
                        </span>
                      ))}
                    </div>
                  </TableCell>
                  <TableCell>{item.quantidade}</TableCell>
                  {!somenteLeitura && (
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          aria-label={`Editar ${item.meta_nome}`}
                          onClick={() => abrirEdicao(item)}
                        >
                          <PencilIcon aria-hidden="true" />
                        </Button>
                        {!item.tem_entrega && (
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            aria-label={`Remover ${item.meta_nome} do plano`}
                            disabled={idExcluindo === item.id}
                            onClick={() => setParaExcluir(item)}
                          >
                            <Trash2Icon aria-hidden="true" />
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  )}
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <p className="text-sm text-muted-foreground">
            Total exigido deste curso: {totalExigido} entregas aceitas
          </p>
          <p className="text-sm text-muted-foreground">
            Uma entrega atende todos os indicadores de cada meta — a quantidade não é multiplicada
            por eles. A quantidade é deste plano; a mesma meta pode exigir outro número em outro
            curso.
          </p>
        </>
      )}

      <ItemPlanoFormModal
        open={modalAberto}
        onOpenChange={setModalAberto}
        planoId={plano.id}
        item={emEdicao}
        onSucesso={(atualizado) => {
          onAtualizado(atualizado)
          setModalAberto(false)
        }}
      />

      <ConfirmDialog
        open={!!paraExcluir}
        onOpenChange={(aberto) => !aberto && setParaExcluir(null)}
        titulo="Remover meta do plano?"
        descricao={`Remover ${paraExcluir?.meta_nome ?? ""} deste plano?`}
        rotuloConfirmar="Remover"
        rotuloConfirmando="Removendo..."
        destrutivo
        confirmando={idExcluindo === paraExcluir?.id}
        onConfirmar={confirmarExclusao}
      />
    </div>
  )
}
