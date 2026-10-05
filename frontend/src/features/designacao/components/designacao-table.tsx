import { PencilIcon, Trash2Icon, XCircleIcon } from "lucide-react"
import Link from "next/link"
import { CabecalhoOrdenavel, type Direcao } from "@/components/shared/ui/cabecalho-ordenavel"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHeader, TableRow } from "@/components/ui/table"
import { formatarDataPura } from "@/lib/formato"
import type { Designacao } from "../types"

function situacaoBadge(situacao: Designacao["situacao"]) {
  if (situacao === "vigente") return { label: "Vigente", variant: "default" as const }
  if (situacao === "futura") return { label: "Futura", variant: "secondary" as const }
  return { label: "Encerrada", variant: "outline" as const }
}

interface DesignacaoTableProps {
  itens: Designacao[]
  cursoId: string
  sort: string
  order: Direcao
  onOrdenar: (campo: string) => void
  onEncerrar: (item: Designacao) => void
  onExcluir: (item: Designacao) => void
  idEmAndamento: string | null
}

export function DesignacaoTable({
  itens,
  cursoId,
  sort,
  order,
  onOrdenar,
  onEncerrar,
  onExcluir,
  idEmAndamento,
}: DesignacaoTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <CabecalhoOrdenavel
            campo="coordenador"
            rotulo="Coordenador"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
          />
          <CabecalhoOrdenavel
            campo="portaria"
            rotulo="Portaria"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
            className="hidden md:table-cell"
          />
          <CabecalhoOrdenavel
            campo="data_inicio"
            rotulo="Início"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
          />
          <CabecalhoOrdenavel
            campo="data_fim"
            rotulo="Fim"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
          />
          <CabecalhoOrdenavel
            campo="situacao"
            rotulo="Situação"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
            ordenavel={false}
          />
          <CabecalhoOrdenavel
            campo="acoes"
            rotulo="Ações"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
            ordenavel={false}
          />
        </TableRow>
      </TableHeader>
      <TableBody>
        {itens.map((item) => {
          const badge = situacaoBadge(item.situacao)
          return (
            <TableRow key={item.id}>
              <TableCell>
                <div>
                  <span>{item.coordenador.nome}</span>
                  {item.autodesignacao && (
                    <span className="block text-xs text-muted-foreground">ⓘ autodesignação</span>
                  )}
                </div>
              </TableCell>
              <TableCell className="hidden md:table-cell">{item.portaria}</TableCell>
              <TableCell>{formatarDataPura(item.data_inicio)}</TableCell>
              <TableCell>{item.data_fim ? formatarDataPura(item.data_fim) : "—"}</TableCell>
              <TableCell>
                <StatusBadge label={badge.label} variant={badge.variant} />
              </TableCell>
              <TableCell>
                <div className="flex items-center gap-1">
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    aria-label={`Editar designação de ${item.coordenador.nome}`}
                    render={<Link href={`/app/cursos/${cursoId}/designacoes/${item.id}`} />}
                  >
                    <PencilIcon aria-hidden="true" />
                    <IndicadorDeNavegacao />
                  </Button>
                  {item.situacao === "vigente" && (
                    <LoadingButton
                      variant="ghost"
                      size="icon-sm"
                      loading={idEmAndamento === item.id}
                      loadingText=""
                      aria-label={`Encerrar designação de ${item.coordenador.nome}`}
                      onClick={() => onEncerrar(item)}
                    >
                      <XCircleIcon aria-hidden="true" />
                    </LoadingButton>
                  )}
                  {item.situacao === "futura" && (
                    <Button
                      variant="ghost"
                      size="icon-sm"
                      aria-label={`Excluir designação de ${item.coordenador.nome}`}
                      onClick={() => onExcluir(item)}
                    >
                      <Trash2Icon aria-hidden="true" />
                    </Button>
                  )}
                </div>
              </TableCell>
            </TableRow>
          )
        })}
      </TableBody>
    </Table>
  )
}
