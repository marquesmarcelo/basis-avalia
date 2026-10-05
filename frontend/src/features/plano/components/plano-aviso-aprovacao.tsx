import { AlertTriangleIcon } from "lucide-react"
import { Alert, AlertDescription } from "@/components/ui/alert"
import type { Plano } from "../types"

interface PlanoAvisoAprovacaoProps {
  plano: Plano
}

export function PlanoAvisoAprovacao({ plano }: PlanoAvisoAprovacaoProps) {
  if (!plano.sem_aprovacao) return null
  const texto =
    plano.situacao === "encerrado"
      ? "Este plano foi encerrado sem aprovação registrada."
      : "Este plano está vigente e ainda não tem aprovação registrada. Informe a data e o órgão quando a reunião acontecer."
  return (
    <Alert role="status">
      <AlertTriangleIcon aria-hidden="true" />
      <AlertDescription>{texto}</AlertDescription>
    </Alert>
  )
}
