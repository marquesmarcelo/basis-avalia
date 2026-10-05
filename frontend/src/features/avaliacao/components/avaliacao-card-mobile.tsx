import { InfoIcon } from "lucide-react"
import Link from "next/link"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Card, CardContent } from "@/components/ui/card"
import type { Entrega } from "@/features/entrega/types"

interface AvaliacaoCardMobileProps {
  item: Entrega
}

function rotuloSituacao(situacao: Entrega["situacao"]) {
  if (situacao === "aceita") return "Aceita"
  if (situacao === "recusada") return "Em correção"
  return "Pendente"
}

export function AvaliacaoCardMobile({ item }: AvaliacaoCardMobileProps) {
  return (
    <Link href={`/app/avaliacoes/${item.id}`} className="block">
      <Card>
        <CardContent className="flex flex-col gap-2">
          <div className="flex items-center justify-between gap-2">
            <p className="font-medium">{item.curso_nome}</p>
            <StatusBadge label={rotuloSituacao(item.situacao)} />
          </div>
          <p className="text-sm">
            {item.meta_nome}
            {item.coordenado_pelo_avaliador && (
              <InfoIcon
                className="ml-1 inline size-3.5 text-amber-600"
                aria-label="Este curso é coordenado por você"
              />
            )}
          </p>
          <p className="text-sm text-muted-foreground">
            {new Date(item.criado_em).toLocaleDateString("pt-BR")}
          </p>
        </CardContent>
      </Card>
    </Link>
  )
}
