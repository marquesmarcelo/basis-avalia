import { BanIcon, PencilIcon, RotateCcwIcon, Trash2Icon, UnlinkIcon } from "lucide-react"
import Link from "next/link"
import { CabecalhoOrdenavel, type Direcao } from "@/components/shared/ui/cabecalho-ordenavel"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHeader, TableRow } from "@/components/ui/table"
import type { Meta } from "../types"

interface MetaTableProps {
  itens: Meta[]
  sort: string
  order: Direcao
  onOrdenar: (campo: string) => void
  onAlterarSituacao: (item: Meta) => void
  onExcluir: (item: Meta) => void
  idAlterandoSituacao: string | null
  voltar?: string
  indicadorContextoId?: string
  onDesvincular?: (item: Meta) => void
  idDesvinculando?: string | null
}

export function MetaTable({
  itens,
  sort,
  order,
  onOrdenar,
  onAlterarSituacao,
  onExcluir,
  idAlterandoSituacao,
  voltar,
  indicadorContextoId,
  onDesvincular,
  idDesvinculando,
}: MetaTableProps) {
  const linkEditar = (id: string) =>
    voltar ? `/app/metas/${id}?voltar=${encodeURIComponent(voltar)}` : `/app/metas/${id}`

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <CabecalhoOrdenavel
            campo="nome"
            rotulo="Nome"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
          />
          <CabecalhoOrdenavel
            campo="indicadores"
            rotulo="Indicadores"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
            ordenavel={false}
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
            campo="planos"
            rotulo="Planos"
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
          const soEsteIndicador =
            !!indicadorContextoId &&
            item.indicadores.length === 1 &&
            item.indicadores[0].id === indicadorContextoId
          return (
            <TableRow key={item.id}>
              <TableCell>{item.nome}</TableCell>
              <TableCell>
                <div className="flex flex-col gap-0.5">
                  {item.indicadores.map((ind) => (
                    <span key={ind.id} className="text-sm">
                      {ind.codigo}
                      {ind.situacao === "inativo" && " (inativo)"} ·{" "}
                      {ind.escopo === "plataforma" ? "Do INEP" : "Próprio"}
                    </span>
                  ))}
                </div>
              </TableCell>
              <TableCell>
                <StatusBadge
                  label={item.situacao === "ativo" ? "Ativa" : "Inativa"}
                  variant={item.situacao === "ativo" ? "default" : "secondary"}
                />
              </TableCell>
              <TableCell className="text-center">{item.planos}</TableCell>
              <TableCell>
                <div className="flex items-center gap-1">
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    render={<Link href={linkEditar(item.id)} aria-label={`Editar ${item.nome}`} />}
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
                      title={
                        soEsteIndicador
                          ? "Único indicador desta meta — não pode ser desvinculada"
                          : "Desvincular deste indicador (a meta continua existindo)"
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
              </TableCell>
            </TableRow>
          )
        })}
      </TableBody>
    </Table>
  )
}
