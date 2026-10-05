import { BanIcon, ListIcon, PencilIcon, RotateCcwIcon, Trash2Icon } from "lucide-react"
import Link from "next/link"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import type { Indicador } from "../types"
import { IndicadorTravado } from "./indicador-travado"

interface IndicadorCardMobileProps {
  item: Indicador
  onAlterarSituacao: (item: Indicador) => void
  onExcluir: (item: Indicador) => void
  idAlterandoSituacao: string | null
}

export function IndicadorCardMobile({
  item,
  onAlterarSituacao,
  onExcluir,
  idAlterandoSituacao,
}: IndicadorCardMobileProps) {
  const proprio = item.escopo === "instituicao"
  return (
    <Card>
      <CardContent className="flex flex-col gap-2">
        <div className="flex items-center justify-between gap-2">
          <div>
            <p className="font-medium">
              <Link href={`/app/indicadores/${item.id}`} className="hover:underline">
                {item.codigo} — {item.nome}
              </Link>
            </p>
            <p className="text-sm text-muted-foreground">{item.metas} meta(s)</p>
          </div>
          <div className="flex flex-col items-end gap-1">
            <StatusBadge
              label={proprio ? "Próprio" : "Do INEP"}
              variant={proprio ? "outline" : "secondary"}
            />
            <StatusBadge
              label={item.situacao === "ativo" ? "Ativo" : "Inativo"}
              variant={item.situacao === "ativo" ? "default" : "secondary"}
            />
          </div>
        </div>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon-sm"
            render={
              <Link
                href={`/app/indicadores/${item.id}`}
                aria-label={proprio ? `Editar ${item.codigo}` : `Ver metas de ${item.codigo}`}
              />
            }
          >
            {proprio ? <PencilIcon aria-hidden="true" /> : <ListIcon aria-hidden="true" />}
            <IndicadorDeNavegacao />
          </Button>
          {proprio ? (
            <>
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
            </>
          ) : (
            <IndicadorTravado />
          )}
        </div>
      </CardContent>
    </Card>
  )
}
