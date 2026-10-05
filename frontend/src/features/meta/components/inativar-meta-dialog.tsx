import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import type { Meta } from "../types"

interface InativarMetaDialogProps {
  meta: Meta | null
  onOpenChange: (open: boolean) => void
  confirmando: boolean
  onConfirmar: () => void
}

export function InativarMetaDialog({
  meta,
  onOpenChange,
  confirmando,
  onConfirmar,
}: InativarMetaDialogProps) {
  return (
    <ConfirmDialog
      open={!!meta}
      onOpenChange={onOpenChange}
      titulo={`Inativar ${meta?.nome ?? ""}?`}
      descricao="Os planos vigentes que já a usam continuam sendo cobrados normalmente. Ela some do autocomplete de novos itens de plano."
      rotuloConfirmar="Inativar"
      rotuloConfirmando="Inativando..."
      destrutivo
      confirmando={confirmando}
      onConfirmar={onConfirmar}
    />
  )
}
