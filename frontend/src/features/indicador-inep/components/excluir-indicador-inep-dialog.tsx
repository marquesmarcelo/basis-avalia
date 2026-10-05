import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import type { IndicadorInep } from "../types"

interface ExcluirIndicadorInepDialogProps {
  indicador: IndicadorInep | null
  onOpenChange: (open: boolean) => void
  confirmando: boolean
  onConfirmar: () => void
}

export function ExcluirIndicadorInepDialog({
  indicador,
  onOpenChange,
  confirmando,
  onConfirmar,
}: ExcluirIndicadorInepDialogProps) {
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
