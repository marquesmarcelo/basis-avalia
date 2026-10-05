import { CabecalhoOrdenavel, type Direcao } from "@/components/shared/ui/cabecalho-ordenavel"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Table, TableBody, TableCell, TableHeader, TableRow } from "@/components/ui/table"
import type { LinhaRelatorio } from "../types"

const rotuloSituacao: Record<string, string> = {
  cumprida: "Cumprida",
  sem_responsavel: "Sem responsável",
  em_andamento: "Em andamento",
  em_correcao: "Em correção",
  nao_cumprida: "Não cumprida",
}

interface RelatorioTableProps {
  itens: LinhaRelatorio[]
  sort: string
  order: Direcao
  onOrdenar: (campo: string) => void
}

export function RelatorioTable({ itens, sort, order, onOrdenar }: RelatorioTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <CabecalhoOrdenavel
            campo="curso"
            rotulo="Curso"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
          />
          <CabecalhoOrdenavel
            campo="responsavel"
            rotulo="Responsável"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
          />
          <CabecalhoOrdenavel
            campo="meta"
            rotulo="Meta"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
          />
          <CabecalhoOrdenavel
            campo="exigido"
            rotulo="Exigido"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
          />
          <CabecalhoOrdenavel
            campo="aceitas"
            rotulo="Aceitas"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
          />
          <CabecalhoOrdenavel
            campo="cumprimento"
            rotulo="Cumprimento"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
          />
          <CabecalhoOrdenavel
            campo=""
            rotulo="Situação"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
            ordenavel={false}
          />
        </TableRow>
      </TableHeader>
      <TableBody>
        {itens.map((l) => (
          <TableRow key={l.item_plano_id}>
            <TableCell>{l.curso_nome}</TableCell>
            <TableCell>
              {l.responsavel_nome ??
                (l.vago_desde ? `Vago desde ${l.vago_desde}` : "Vago o período inteiro")}
            </TableCell>
            <TableCell>
              {l.meta_nome}
              {l.avaliacao_pelo_proprio_coordenador && (
                <span className="ml-1 text-xs text-amber-600">
                  (avaliação pelo próprio coordenador)
                </span>
              )}
            </TableCell>
            <TableCell>{l.exigido}</TableCell>
            <TableCell>{l.aceitas}</TableCell>
            <TableCell>{l.cumprimento.toFixed(0)}%</TableCell>
            <TableCell>
              <StatusBadge label={rotuloSituacao[l.situacao] ?? l.situacao} />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
