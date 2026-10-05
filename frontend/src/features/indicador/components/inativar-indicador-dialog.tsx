import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import type { Indicador } from "../types"

interface InativarIndicadorDialogProps {
  indicador: Indicador | null
  onOpenChange: (open: boolean) => void
  confirmando: boolean
  onConfirmar: () => void
}

export function InativarIndicadorDialog({
  indicador,
  onOpenChange,
  confirmando,
  onConfirmar,
}: InativarIndicadorDialogProps) {
  return (
    <ConfirmDialog
      open={!!indicador}
      onOpenChange={onOpenChange}
      titulo={`Inativar ${indicador?.codigo ?? ""} — ${indicador?.nome ?? ""}?`}
      descricao="Ele some do autocomplete de nova meta. As metas que já o referenciam continuam funcionando."
      rotuloConfirmar="Inativar"
      rotuloConfirmando="Inativando..."
      destrutivo
      confirmando={confirmando}
      onConfirmar={onConfirmar}
    />
  )
}
