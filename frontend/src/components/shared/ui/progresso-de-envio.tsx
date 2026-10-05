import { Progress } from "@/components/ui/progress"

interface ProgressoDeEnvioProps {
  percentual: number
  indeterminado?: boolean
}

export function ProgressoDeEnvio({ percentual, indeterminado = false }: ProgressoDeEnvioProps) {
  return (
    <div role="group" aria-label="Progresso do envio" className="flex items-center gap-2">
      <Progress
        value={indeterminado ? null : percentual}
        aria-valuetext={indeterminado ? "Enviando" : `${percentual}% enviado`}
        className="flex-1"
      />
      <span aria-hidden="true" className="w-12 text-right text-xs text-muted-foreground">
        {indeterminado ? "..." : `${percentual}%`}
      </span>
    </div>
  )
}
