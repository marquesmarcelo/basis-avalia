import { BanIcon, PencilIcon, RotateCcwIcon, Trash2Icon, UsersIcon } from "lucide-react"
import Link from "next/link"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import type { Curso } from "../types"
import { CursoCoordenadorCelula } from "./curso-coordenador-celula"

interface CursoCardMobileProps {
  item: Curso
  onVerDesignacoes: (item: Curso) => void
  onAlterarSituacao: (item: Curso) => void
  onExcluir: (item: Curso) => void
  idAlterandoSituacao: string | null
}

export function CursoCardMobile({
  item,
  onVerDesignacoes,
  onAlterarSituacao,
  onExcluir,
  idAlterandoSituacao,
}: CursoCardMobileProps) {
  return (
    <Card>
      <CardContent className="flex flex-col gap-2">
        <div className="flex items-center justify-between gap-2">
          <p className="font-medium">{item.nome}</p>
          <StatusBadge
            label={item.situacao === "ativo" ? "Ativo" : "Inativo"}
            variant={item.situacao === "ativo" ? "default" : "secondary"}
          />
        </div>
        <CursoCoordenadorCelula coordenador={item.coordenador} />
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon-sm"
            render={<Link href={`/app/cursos/${item.id}`} aria-label={`Editar ${item.nome}`} />}
          >
            <PencilIcon aria-hidden="true" />
            <IndicadorDeNavegacao />
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={`Ver designações de ${item.nome}`}
            onClick={() => onVerDesignacoes(item)}
          >
            <UsersIcon aria-hidden="true" />
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
          {!item.tem_vinculo && (
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
