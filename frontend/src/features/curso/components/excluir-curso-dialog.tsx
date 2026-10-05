import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import type { Curso } from "../types"

interface ExcluirCursoDialogProps {
  curso: Curso | null
  onOpenChange: (open: boolean) => void
  confirmando: boolean
  onConfirmar: () => void
}

export function ExcluirCursoDialog({
  curso,
  onOpenChange,
  confirmando,
  onConfirmar,
}: ExcluirCursoDialogProps) {
  return (
    <ConfirmDialog
      open={!!curso}
      onOpenChange={onOpenChange}
      titulo={`Excluir ${curso?.nome ?? ""}?`}
      descricao="Esta ação não pode ser desfeita."
      rotuloConfirmar="Excluir"
      rotuloConfirmando="Excluindo..."
      destrutivo
      confirmando={confirmando}
      onConfirmar={onConfirmar}
    />
  )
}
