import { PencilIcon, UserPlusIcon } from "lucide-react"
import Link from "next/link"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import type { Instituicao } from "../types"

interface InstituicaoCardMobileProps {
  item: Instituicao
  onAlterarSituacao: (item: Instituicao) => void
  idAlterandoSituacao: string | null
}

export function InstituicaoCardMobile({
  item,
  onAlterarSituacao,
  idAlterandoSituacao,
}: InstituicaoCardMobileProps) {
  return (
    <Card>
      <CardContent className="flex flex-col gap-2">
        <div className="flex items-center justify-between gap-2">
          <div>
            <Link
              href={`/app/instituicoes/${item.id}/pesquisadores`}
              className="font-medium hover:underline"
            >
              {item.nome}
            </Link>
            <p className="text-sm text-muted-foreground">{item.sigla}</p>
          </div>
          <StatusBadge
            label={item.situacao === "ativa" ? "Ativa" : "Inativa"}
            variant={item.situacao === "ativa" ? "default" : "secondary"}
          />
        </div>
        {item.pesquisadores_ativos === 0 && (
          <StatusBadge
            label="Sem Pesquisador Institucional"
            variant="destructive"
            className="w-fit"
          />
        )}
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon-sm"
            render={
              <Link href={`/app/instituicoes/${item.id}`} aria-label={`Editar ${item.nome}`} />
            }
          >
            <PencilIcon aria-hidden="true" />
            <IndicadorDeNavegacao />
          </Button>
          {item.pesquisadores_ativos === 0 && (
            <Button
              variant="outline"
              size="icon-sm"
              render={
                <Link
                  href={`/app/instituicoes/${item.id}/pesquisadores/novo`}
                  aria-label={`Cadastrar Pesquisador Institucional de ${item.nome}`}
                />
              }
            >
              <UserPlusIcon aria-hidden="true" />
              <IndicadorDeNavegacao />
            </Button>
          )}
          <LoadingButton
            variant="outline"
            size="sm"
            loading={idAlterandoSituacao === item.id}
            loadingText={item.situacao === "ativa" ? "Inativando..." : "Ativando..."}
            aria-label={item.situacao === "ativa" ? `Inativar ${item.nome}` : `Ativar ${item.nome}`}
            onClick={() => onAlterarSituacao(item)}
          >
            {item.situacao === "ativa" ? "Inativar" : "Ativar"}
          </LoadingButton>
        </div>
      </CardContent>
    </Card>
  )
}
