import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import type { Meta } from "../types"

interface ExcluirMetaDialogProps {
  meta: Meta | null
  onOpenChange: (open: boolean) => void
  confirmando: boolean
  onConfirmar: () => void
}

export function ExcluirMetaDialog({
  meta,
  onOpenChange,
  confirmando,
  onConfirmar,
}: ExcluirMetaDialogProps) {
  return (
    <ConfirmDialog
      open={!!meta}
      onOpenChange={onOpenChange}
      titulo={`Excluir ${meta?.nome ?? ""}?`}
      descricao="Esta ação não pode ser desfeita."
      rotuloConfirmar="Excluir"
      rotuloConfirmando="Excluindo..."
      destrutivo
      confirmando={confirmando}
      onConfirmar={onConfirmar}
    />
  )
}
