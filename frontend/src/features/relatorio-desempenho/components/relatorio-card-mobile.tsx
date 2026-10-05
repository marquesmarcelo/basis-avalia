import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Card, CardContent } from "@/components/ui/card"
import type { LinhaRelatorio } from "../types"

const rotuloSituacao: Record<string, string> = {
  cumprida: "Cumprida",
  sem_responsavel: "Sem responsável",
  em_andamento: "Em andamento",
  em_correcao: "Em correção",
  nao_cumprida: "Não cumprida",
}

interface RelatorioCardMobileProps {
  item: LinhaRelatorio
}

export function RelatorioCardMobile({ item }: RelatorioCardMobileProps) {
  return (
    <Card>
      <CardContent className="flex flex-col gap-2">
        <div className="flex items-center justify-between gap-2">
          <p className="font-medium">{item.curso_nome}</p>
          <StatusBadge label={rotuloSituacao[item.situacao] ?? item.situacao} />
        </div>
        <p className="text-sm">
          {item.meta_nome}
          {item.avaliacao_pelo_proprio_coordenador && (
            <span className="ml-1 text-xs text-amber-600">
              (avaliação pelo próprio coordenador)
            </span>
          )}
        </p>
        <p className="text-sm text-muted-foreground">
          {item.responsavel_nome ??
            (item.vago_desde ? `Vago desde ${item.vago_desde}` : "Vago o período inteiro")}
        </p>
        <p className="text-sm text-muted-foreground">
          Exigido {item.exigido} · Aceitas {item.aceitas} · Cumprimento{" "}
          {item.cumprimento.toFixed(0)}%
        </p>
      </CardContent>
    </Card>
  )
}
