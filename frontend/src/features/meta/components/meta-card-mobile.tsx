import { BanIcon, PencilIcon, RotateCcwIcon, Trash2Icon, UnlinkIcon } from "lucide-react"
import Link from "next/link"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import type { Meta } from "../types"

interface MetaCardMobileProps {
  item: Meta
  onAlterarSituacao: (item: Meta) => void
  onExcluir: (item: Meta) => void
  idAlterandoSituacao: string | null
  voltar?: string
  indicadorContextoId?: string
  onDesvincular?: (item: Meta) => void
  idDesvinculando?: string | null
}

export function MetaCardMobile({
  item,
  onAlterarSituacao,
  onExcluir,
  idAlterandoSituacao,
  voltar,
  indicadorContextoId,
  onDesvincular,
  idDesvinculando,
}: MetaCardMobileProps) {
  const soEsteIndicador =
    !!indicadorContextoId &&
    item.indicadores.length === 1 &&
    item.indicadores[0].id === indicadorContextoId
  const linkEditar = voltar
    ? `/app/metas/${item.id}?voltar=${encodeURIComponent(voltar)}`
    : `/app/metas/${item.id}`

  return (
    <Card>
      <CardContent className="flex flex-col gap-2">
        <div className="flex items-center justify-between gap-2">
          <p className="font-medium">{item.nome}</p>
          <StatusBadge
            label={item.situacao === "ativo" ? "Ativa" : "Inativa"}
            variant={item.situacao === "ativo" ? "default" : "secondary"}
          />
        </div>
        <div className="flex flex-wrap gap-1">
          {item.indicadores.map((ind) => (
            <StatusBadge
              key={ind.id}
              label={`${ind.codigo}${ind.situacao === "inativo" ? " (inativo)" : ""}`}
              variant={ind.escopo === "plataforma" ? "secondary" : "outline"}
            />
          ))}
        </div>
        <p className="text-sm text-muted-foreground">{item.planos} plano(s)</p>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon-sm"
            render={<Link href={linkEditar} aria-label={`Editar ${item.nome}`} />}
          >
            <PencilIcon aria-hidden="true" />
            <IndicadorDeNavegacao />
          </Button>
          <LoadingButton
            variant="ghost"
            size="icon-sm"
            loading={idAlterandoSituacao === item.id}
            loadingText=""
            aria-label={
              item.situacao === "ativo" ? `Inativar ${item.nome}` : `Reativar ${item.nome}`
            }
            onClick={() => onAlterarSituacao(item)}
          >
            {item.situacao === "ativo" ? (
              <BanIcon aria-hidden="true" />
            ) : (
              <RotateCcwIcon aria-hidden="true" />
            )}
          </LoadingButton>
          {onDesvincular && (
            <LoadingButton
              variant="ghost"
              size="icon-sm"
              loading={idDesvinculando === item.id}
              loadingText=""
              disabled={soEsteIndicador}
              aria-label={
                soEsteIndicador
                  ? `${item.nome} não pode ser desvinculada — é o único indicador dela`
                  : `Desvincular ${item.nome} deste indicador`
              }
              onClick={() => onDesvincular(item)}
            >
              <UnlinkIcon aria-hidden="true" />
            </LoadingButton>
          )}
          {item.planos === 0 && (
            <Button
              variant="ghost"
              size="icon-sm"
              aria-label={`Excluir ${item.nome}`}
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
