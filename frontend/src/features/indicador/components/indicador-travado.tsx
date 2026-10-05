import { LockIcon } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"

export function IndicadorTravado() {
  return (
    <Tooltip>
      <TooltipTrigger
        render={<Badge variant="outline" tabIndex={0} className="gap-1 text-muted-foreground" />}
      >
        <LockIcon aria-hidden="true" />
        <span className="text-xs">Somente leitura</span>
      </TooltipTrigger>
      <TooltipContent>
        Mantido pelo Administrador do Sistema. Você pode usar este indicador em metas normalmente.
      </TooltipContent>
    </Tooltip>
  )
}
