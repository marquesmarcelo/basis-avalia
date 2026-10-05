import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import type { Indicador } from "../types"

interface ExcluirIndicadorDialogProps {
  indicador: Indicador | null
  onOpenChange: (open: boolean) => void
  confirmando: boolean
  onConfirmar: () => void
}

export function ExcluirIndicadorDialog({
  indicador,
  onOpenChange,
  confirmando,
  onConfirmar,
}: ExcluirIndicadorDialogProps) {
  return (
    <ConfirmDialog
      open={!!indicador}
      onOpenChange={onOpenChange}
      titulo={`Excluir ${indicador?.codigo ?? ""} — ${indicador?.nome ?? ""}?`}
      descricao="Esta ação não pode ser desfeita."
      rotuloConfirmar="Excluir"
      rotuloConfirmando="Excluindo..."
      destrutivo
      confirmando={confirmando}
      onConfirmar={onConfirmar}
    />
  )
}
