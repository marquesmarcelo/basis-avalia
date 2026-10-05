import { CopyIcon, DownloadIcon, Trash2Icon } from "lucide-react"
import Link from "next/link"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import type { Plano } from "../types"

const ROTULO_SITUACAO: Record<Plano["situacao"], string> = {
  rascunho: "Rascunho",
  vigente: "Vigente",
  encerrado: "Encerrado",
}

interface PlanoCardMobileProps {
  item: Plano
  somenteLeitura?: boolean
  onExcluir: (item: Plano) => void
  onGerarDocumento: (item: Plano) => void
  idExcluindo: string | null
  idGerandoDocumento: string | null
}

export function PlanoCardMobile({
  item,
  somenteLeitura,
  onExcluir,
  onGerarDocumento,
  idExcluindo,
  idGerandoDocumento,
}: PlanoCardMobileProps) {
  return (
    <Card>
      <CardContent className="flex flex-col gap-2">
        <div className="flex items-center justify-between gap-2">
          <Link
            href={`/app/planos/${item.id}`}
            className="font-medium underline-offset-2 hover:underline"
          >
            {item.curso.nome}
          </Link>
          <StatusBadge
            label={ROTULO_SITUACAO[item.situacao]}
            variant={item.situacao === "vigente" ? "default" : "secondary"}
          />
        </div>
        <p className="text-sm text-muted-foreground">
          {item.periodo.nome} · Metas: {item.metas} · Exigido: {item.total_exigido}
        </p>
        {item.sem_aprovacao && <p className="text-sm">⚠ sem aprovação</p>}
        {item.curso.vago && <p className="text-sm text-muted-foreground">Vago ⚠</p>}
        <div className="flex items-center gap-1">
          {!somenteLeitura && (
            <Button
              variant="ghost"
              size="icon-sm"
              aria-label={`Copiar plano de ${item.curso.nome}`}
              render={<Link href={`/app/planos/${item.id}/copiar`} />}
            >
              <CopyIcon aria-hidden="true" />
            </Button>
          )}
          <LoadingButton
            variant="ghost"
            size="icon-sm"
            loading={idGerandoDocumento === item.id}
            loadingText=""
            aria-label={`Baixar documento de ${item.curso.nome}`}
            onClick={() => onGerarDocumento(item)}
          >
            <DownloadIcon aria-hidden="true" />
          </LoadingButton>
          {!somenteLeitura && item.situacao === "rascunho" && (
            <Button
              variant="ghost"
              size="icon-sm"
              aria-label={`Excluir plano de ${item.curso.nome}`}
              disabled={idExcluindo === item.id}
              onClick={() => onExcluir(item)}
            >
              <Trash2Icon aria-hidden="true" />
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
