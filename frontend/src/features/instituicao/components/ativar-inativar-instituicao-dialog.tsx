import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import type { Instituicao } from "../types"

interface AtivarInativarInstituicaoDialogProps {
  instituicao: Instituicao | null
  onOpenChange: (open: boolean) => void
  confirmando: boolean
  onConfirmar: () => void
}

export function AtivarInativarInstituicaoDialog({
  instituicao,
  onOpenChange,
  confirmando,
  onConfirmar,
}: AtivarInativarInstituicaoDialogProps) {
  return (
    <ConfirmDialog
      open={!!instituicao}
      onOpenChange={onOpenChange}
      titulo={`Inativar ${instituicao?.nome ?? ""}?`}
      descricao="Todas as pessoas desta instituição perderão o acesso — na próxima ação que realizarem, serão desconectadas e não conseguirão entrar novamente até que a instituição seja reativada."
      rotuloConfirmar="Inativar"
      rotuloConfirmando="Inativando..."
      destrutivo
      confirmando={confirmando}
      onConfirmar={onConfirmar}
    />
  )
}
