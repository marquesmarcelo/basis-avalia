import { BanIcon, ListIcon, PencilIcon, RotateCcwIcon, Trash2Icon } from "lucide-react"
import Link from "next/link"
import { CabecalhoOrdenavel, type Direcao } from "@/components/shared/ui/cabecalho-ordenavel"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHeader, TableRow } from "@/components/ui/table"
import type { Indicador } from "../types"
import { IndicadorTravado } from "./indicador-travado"

interface IndicadorTableProps {
  itens: Indicador[]
  sort: string
  order: Direcao
  onOrdenar: (campo: string) => void
  onAlterarSituacao: (item: Indicador) => void
  onExcluir: (item: Indicador) => void
  idAlterandoSituacao: string | null
}

export function IndicadorTable({
  itens,
  sort,
  order,
  onOrdenar,
  onAlterarSituacao,
  onExcluir,
  idAlterandoSituacao,
}: IndicadorTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <CabecalhoOrdenavel
            campo="codigo"
            rotulo="Código"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
          />
          <CabecalhoOrdenavel
            campo="nome"
            rotulo="Nome"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
          />
          <CabecalhoOrdenavel
            campo="escopo"
            rotulo="Origem"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
          />
          <CabecalhoOrdenavel
            campo="referencia"
            rotulo="Referência"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
            ordenavel={false}
            className="hidden lg:table-cell"
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
            campo="metas"
            rotulo="Metas"
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
          const proprio = item.escopo === "instituicao"
          return (
            <TableRow key={item.id}>
              <TableCell>{item.codigo}</TableCell>
              <TableCell>
                <Link href={`/app/indicadores/${item.id}`} className="hover:underline">
                  {item.nome}
                </Link>
              </TableCell>
              <TableCell>
                <StatusBadge
                  label={proprio ? "Próprio" : "Do INEP"}
                  variant={proprio ? "outline" : "secondary"}
                />
              </TableCell>
              <TableCell className="hidden max-w-xs truncate lg:table-cell">
                {item.referencia_instrumento ?? "—"}
              </TableCell>
              <TableCell>
                <StatusBadge
                  label={item.situacao === "ativo" ? "Ativo" : "Inativo"}
                  variant={item.situacao === "ativo" ? "default" : "secondary"}
                />
              </TableCell>
              <TableCell className="text-center">{item.metas}</TableCell>
              <TableCell>
                <div className="flex items-center gap-1">
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    render={
                      <Link
                        href={`/app/indicadores/${item.id}`}
                        aria-label={
                          proprio ? `Editar ${item.codigo}` : `Ver metas de ${item.codigo}`
                        }
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
                          item.situacao === "ativo"
                            ? `Inativar ${item.codigo}`
                            : `Reativar ${item.codigo}`
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
              </TableCell>
            </TableRow>
          )
        })}
      </TableBody>
    </Table>
  )
}
