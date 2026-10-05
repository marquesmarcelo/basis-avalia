import { BanIcon, PencilIcon, RotateCcwIcon, Trash2Icon } from "lucide-react"
import Link from "next/link"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import type { IndicadorInep } from "../types"

interface IndicadorInepCardMobileProps {
  item: IndicadorInep
  onAlterarSituacao: (item: IndicadorInep) => void
  onExcluir: (item: IndicadorInep) => void
  idAlterandoSituacao: string | null
}

export function IndicadorInepCardMobile({
  item,
  onAlterarSituacao,
  onExcluir,
  idAlterandoSituacao,
}: IndicadorInepCardMobileProps) {
  return (
    <Card>
      <CardContent className="flex flex-col gap-2">
        <div className="flex items-center justify-between gap-2">
          <div>
            <p className="font-medium">
              {item.codigo} — {item.nome}
            </p>
            <p className="text-sm text-muted-foreground">{item.metas} meta(s)</p>
          </div>
          <StatusBadge
            label={item.situacao === "ativo" ? "Ativo" : "Inativo"}
            variant={item.situacao === "ativo" ? "default" : "secondary"}
          />
        </div>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon-sm"
            render={
              <Link href={`/app/indicadores-inep/${item.id}`} aria-label={`Editar ${item.codigo}`} />
            }
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
              item.situacao === "ativo" ? `Inativar ${item.codigo}` : `Reativar ${item.codigo}`
            }
            onClick={() => onAlterarSituacao(item)}
          >
            {item.situacao === "ativo" ? (
              <BanIcon aria-hidden="true" />
            ) : (
              <RotateCcwIcon aria-hidden="true" />
            )}
          </LoadingButton>
          {item.metas === 0 && (
            <Button
              variant="ghost"
              size="icon-sm"
              aria-label={`Excluir ${item.codigo}`}
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
