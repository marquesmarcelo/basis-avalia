import { PencilIcon, Trash2Icon } from "lucide-react"
import Link from "next/link"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { formatarDataPura } from "@/lib/formato"
import { ROTULO_SITUACAO } from "../lib/situacao"
import type { Periodo } from "../types"

interface PeriodoCardMobileProps {
  item: Periodo
  onExcluir: (item: Periodo) => void
  idExcluindo: string | null
}

export function PeriodoCardMobile({ item, onExcluir, idExcluindo }: PeriodoCardMobileProps) {
  return (
    <Card>
      <CardContent className="flex flex-col gap-2">
        <div className="flex items-center justify-between gap-2">
          <p className="font-medium">{item.nome}</p>
          <StatusBadge
            label={ROTULO_SITUACAO[item.situacao]}
            variant={item.situacao === "aberto" ? "default" : "secondary"}
          />
        </div>
        <p className="text-sm text-muted-foreground">
          {formatarDataPura(item.data_inicio)} a {formatarDataPura(item.data_fim)} · {item.planos}{" "}
          plano(s)
        </p>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon-sm"
            render={<Link href={`/app/periodos/${item.id}`} aria-label={`Editar ${item.nome}`} />}
          >
            <PencilIcon aria-hidden="true" />
            <IndicadorDeNavegacao />
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={`Excluir ${item.nome}`}
            disabled={idExcluindo === item.id}
            onClick={() => onExcluir(item)}
          >
            <Trash2Icon aria-hidden="true" />
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
