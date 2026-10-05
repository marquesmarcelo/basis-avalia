import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import type { Curso } from "../types"

interface InativarCursoDialogProps {
  curso: Curso | null
  onOpenChange: (open: boolean) => void
  confirmando: boolean
  onConfirmar: () => void
}

export function InativarCursoDialog({
  curso,
  onOpenChange,
  confirmando,
  onConfirmar,
}: InativarCursoDialogProps) {
  return (
    <ConfirmDialog
      open={!!curso}
      onOpenChange={onOpenChange}
      titulo={`Inativar ${curso?.nome ?? ""}?`}
      descricao="O curso sai da cobrança de novos relatórios e não aceita novas entregas. Nada é apagado, e a designação vigente continua ativa — inativar o curso não encerra a portaria do coordenador."
      rotuloConfirmar="Inativar"
      rotuloConfirmando="Inativando..."
      destrutivo
      confirmando={confirmando}
      onConfirmar={onConfirmar}
    />
  )
}
