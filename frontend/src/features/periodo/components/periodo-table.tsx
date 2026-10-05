import { PencilIcon, Trash2Icon } from "lucide-react"
import Link from "next/link"
import { CabecalhoOrdenavel, type Direcao } from "@/components/shared/ui/cabecalho-ordenavel"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHeader, TableRow } from "@/components/ui/table"
import { formatarDataPura } from "@/lib/formato"
import { ROTULO_SITUACAO } from "../lib/situacao"
import type { Periodo } from "../types"

interface PeriodoTableProps {
  itens: Periodo[]
  sort: string
  order: Direcao
  onOrdenar: (campo: string) => void
  onExcluir: (item: Periodo) => void
  idExcluindo: string | null
}

export function PeriodoTable({
  itens,
  sort,
  order,
  onOrdenar,
  onExcluir,
  idExcluindo,
}: PeriodoTableProps) {
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
        {itens.map((item) => (
          <TableRow key={item.id}>
            <TableCell>{item.nome}</TableCell>
            <TableCell>{formatarDataPura(item.data_inicio)}</TableCell>
            <TableCell>{formatarDataPura(item.data_fim)}</TableCell>
            <TableCell>
              <StatusBadge
                label={ROTULO_SITUACAO[item.situacao]}
                variant={item.situacao === "aberto" ? "default" : "secondary"}
              />
            </TableCell>
            <TableCell className="text-center">
              <Button
                variant="link"
                size="sm"
                render={<Link href={`/app/planos?periodo_id=${item.id}`} />}
              >
                {item.planos}
              </Button>
            </TableCell>
            <TableCell>
              <div className="flex items-center gap-1">
                <Button
                  variant="ghost"
                  size="icon-sm"
                  render={
                    <Link href={`/app/periodos/${item.id}`} aria-label={`Editar ${item.nome}`} />
                  }
                >
                  <PencilIcon aria-hidden="true" />
                  <IndicadorDeNavegacao />
                </Button>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  aria-label={`Excluir ${item.nome}`}
                  disabled={idExcluindo === item.id}
                  onClick={() => onExcluir(item)}
                >
                  <Trash2Icon aria-hidden="true" />
                </Button>
              </div>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
