import { TriangleAlertIcon } from "lucide-react"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { formatarDataPura } from "@/lib/formato"
import type { CoordenadorDoCurso } from "../types"

interface CursoCoordenadorCelulaProps {
  coordenador: CoordenadorDoCurso | null
}

export function CursoCoordenadorCelula({ coordenador }: CursoCoordenadorCelulaProps) {
  if (!coordenador) {
    return (
      <span className="inline-flex items-center gap-1 text-amber-700">
        <TriangleAlertIcon className="size-4" aria-hidden="true" />
        Vago
      </span>
    )
  }
  return (
    <div>
      <span>{coordenador.nome}</span>
      {coordenador.data_fim && (
        <span className="block text-xs text-muted-foreground">
          até {formatarDataPura(coordenador.data_fim)}
        </span>
      )}
      {coordenador.tambem_pesquisador_institucional && (
        <Tooltip>
          <TooltipTrigger render={<button type="button" className="ml-1 text-xs" />}>
            também é PI ⓘ
          </TooltipTrigger>
          <TooltipContent>
            Quem coordena este curso também avalia as entregas dele. É permitido, e aparece marcado
            no relatório de desempenho.
          </TooltipContent>
        </Tooltip>
      )}
    </div>
  )
}
