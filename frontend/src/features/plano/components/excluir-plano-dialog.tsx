"use client"

import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import type { Plano } from "../types"

interface ExcluirPlanoDialogProps {
  plano: Plano | null
  onOpenChange: (open: boolean) => void
  confirmando: boolean
  onConfirmar: () => void
}

export function ExcluirPlanoDialog({
  plano,
  onOpenChange,
  confirmando,
  onConfirmar,
}: ExcluirPlanoDialogProps) {
  return (
    <ConfirmDialog
      open={!!plano}
      onOpenChange={onOpenChange}
      titulo="Excluir plano?"
      descricao={`Excluir o plano de ${plano?.curso.nome ?? ""} em ${plano?.periodo.nome ?? ""}? Esta ação não pode ser desfeita.`}
      rotuloConfirmar="Excluir"
      rotuloConfirmando="Excluindo..."
      destrutivo
      confirmando={confirmando}
      onConfirmar={onConfirmar}
    />
  )
}
