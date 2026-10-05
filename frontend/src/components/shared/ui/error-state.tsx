import { AlertTriangleIcon } from "lucide-react"
import { LoadingButton } from "@/components/shared/ui/loading-button"

interface ErrorStateProps {
  titulo: string
  onTentarNovamente: () => void
  tentandoNovamente?: boolean
  rotuloAcao?: string
}

export function ErrorState({
  titulo,
  onTentarNovamente,
  tentandoNovamente = false,
  rotuloAcao = "Tentar novamente",
}: ErrorStateProps) {
  return (
    <div
      className="flex flex-col items-center justify-center gap-3 rounded-lg border border-dashed p-10 text-center"
      role="alert"
    >
      <AlertTriangleIcon className="size-8 text-destructive" aria-hidden="true" />
      <p className="text-sm font-medium">{titulo}</p>
      <LoadingButton
        variant="outline"
        loading={tentandoNovamente}
        loadingText="Tentando novamente..."
        onClick={onTentarNovamente}
      >
        {rotuloAcao}
      </LoadingButton>
    </div>
  )
}
