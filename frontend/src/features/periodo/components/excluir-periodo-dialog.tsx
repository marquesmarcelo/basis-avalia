"use client"

import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import type { Periodo } from "../types"

interface ExcluirPeriodoDialogProps {
  periodo: Periodo | null
  onOpenChange: (open: boolean) => void
  confirmando: boolean
  onConfirmar: () => void
}

export function ExcluirPeriodoDialog({
  periodo,
  onOpenChange,
  confirmando,
  onConfirmar,
}: ExcluirPeriodoDialogProps) {
  return (
    <ConfirmDialog
      open={!!periodo}
      onOpenChange={onOpenChange}
      titulo="Excluir período?"
      descricao={`Excluir o período ${periodo?.nome ?? ""}? Esta ação não pode ser desfeita.`}
      rotuloConfirmar="Excluir"
      rotuloConfirmando="Excluindo..."
      destrutivo
      confirmando={confirmando}
      onConfirmar={onConfirmar}
    />
  )
}
