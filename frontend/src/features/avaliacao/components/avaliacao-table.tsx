import { InfoIcon } from "lucide-react"
import Link from "next/link"
import { CabecalhoOrdenavel, type Direcao } from "@/components/shared/ui/cabecalho-ordenavel"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Table, TableBody, TableCell, TableHeader, TableRow } from "@/components/ui/table"
import type { Entrega } from "@/features/entrega/types"

interface AvaliacaoTableProps {
  itens: Entrega[]
  sort: string
  order: Direcao
  onOrdenar: (campo: string) => void
}

function rotuloSituacao(situacao: Entrega["situacao"]) {
  if (situacao === "aceita") return "Aceita"
  if (situacao === "recusada") return "Em correção"
  return "Pendente"
}

export function AvaliacaoTable({ itens, sort, order, onOrdenar }: AvaliacaoTableProps) {
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
            campo="meta"
            rotulo="Meta"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
          />
          <CabecalhoOrdenavel
            campo="criado_em"
            rotulo="Data"
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
          <CabecalhoOrdenavel
            campo=""
            rotulo=""
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
            ordenavel={false}
          />
        </TableRow>
      </TableHeader>
      <TableBody>
        {itens.map((e) => (
          <TableRow key={e.id}>
            <TableCell>{e.curso_nome}</TableCell>
            <TableCell>
              {e.meta_nome}
              {e.coordenado_pelo_avaliador && (
                <InfoIcon
                  className="ml-1 inline size-3.5 text-amber-600"
                  aria-label="Este curso é coordenado por você"
                />
              )}
            </TableCell>
            <TableCell>{new Date(e.criado_em).toLocaleDateString("pt-BR")}</TableCell>
            <TableCell>
              <StatusBadge label={rotuloSituacao(e.situacao)} />
            </TableCell>
            <TableCell>
              <Link
                href={`/app/avaliacoes/${e.id}`}
                className="text-sm font-medium text-primary hover:underline"
              >
                Avaliar
              </Link>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
