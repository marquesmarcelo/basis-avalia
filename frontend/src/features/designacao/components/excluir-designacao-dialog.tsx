import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import type { Designacao } from "../types"

interface ExcluirDesignacaoDialogProps {
  designacao: Designacao | null
  onOpenChange: (open: boolean) => void
  confirmando: boolean
  onConfirmar: () => void
}

export function ExcluirDesignacaoDialog({
  designacao,
  onOpenChange,
  confirmando,
  onConfirmar,
}: ExcluirDesignacaoDialogProps) {
  return (
    <ConfirmDialog
      open={!!designacao}
      onOpenChange={onOpenChange}
      titulo={`Excluir a designação futura de ${designacao?.coordenador.nome ?? ""}?`}
      descricao="Esta ação não pode ser desfeita."
      rotuloConfirmar="Excluir"
      rotuloConfirmando="Excluindo..."
      destrutivo
      confirmando={confirmando}
      onConfirmar={onConfirmar}
    />
  )
}
