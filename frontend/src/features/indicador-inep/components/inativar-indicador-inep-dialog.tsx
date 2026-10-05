import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import type { IndicadorInep } from "../types"

interface InativarIndicadorInepDialogProps {
  indicador: IndicadorInep | null
  onOpenChange: (open: boolean) => void
  confirmando: boolean
  onConfirmar: () => void
}

export function InativarIndicadorInepDialog({
  indicador,
  onOpenChange,
  confirmando,
  onConfirmar,
}: InativarIndicadorInepDialogProps) {
  return (
    <ConfirmDialog
      open={!!indicador}
      onOpenChange={onOpenChange}
      titulo={`Inativar ${indicador?.codigo ?? ""} — ${indicador?.nome ?? ""}?`}
      descricao="Ele some do autocomplete de todas as instituições. As metas que já o referenciam continuam funcionando e continuam contando."
      rotuloConfirmar="Inativar"
      rotuloConfirmando="Inativando..."
      destrutivo
      confirmando={confirmando}
      onConfirmar={onConfirmar}
    />
  )
}
