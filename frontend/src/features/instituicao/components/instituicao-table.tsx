import { PencilIcon, UserPlusIcon } from "lucide-react"
import Link from "next/link"
import { CabecalhoOrdenavel, type Direcao } from "@/components/shared/ui/cabecalho-ordenavel"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHeader, TableRow } from "@/components/ui/table"
import { formatarData } from "@/lib/formato"
import type { Instituicao } from "../types"

interface InstituicaoTableProps {
  itens: Instituicao[]
  sort: string
  order: Direcao
  onOrdenar: (campo: string) => void
  onAlterarSituacao: (item: Instituicao) => void
  idAlterandoSituacao: string | null
}

export function InstituicaoTable({
  itens,
  sort,
  order,
  onOrdenar,
  onAlterarSituacao,
  idAlterandoSituacao,
}: InstituicaoTableProps) {
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
            campo="sigla"
            rotulo="Sigla"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
          />
          <CabecalhoOrdenavel
            campo="codigo_emec"
            rotulo="Código e-MEC"
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
            campo="criado_em"
            rotulo="Cadastrado em"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
            className="hidden lg:table-cell"
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
            <TableCell>
              <div className="flex items-center gap-2">
                <Link
                  href={`/app/instituicoes/${item.id}/pesquisadores`}
                  className="hover:underline"
                >
                  {item.nome}
                </Link>
                {item.pesquisadores_ativos === 0 && (
                  <StatusBadge label="Sem Pesquisador Institucional" variant="destructive" />
                )}
              </div>
            </TableCell>
            <TableCell>{item.sigla}</TableCell>
            <TableCell className="hidden lg:table-cell">{item.codigo_emec ?? "—"}</TableCell>
            <TableCell>
              <StatusBadge
                label={item.situacao === "ativa" ? "Ativa" : "Inativa"}
                variant={item.situacao === "ativa" ? "default" : "secondary"}
              />
            </TableCell>
            <TableCell className="hidden lg:table-cell">{formatarData(item.criado_em)}</TableCell>
            <TableCell>
              <div className="flex items-center gap-1">
                <Button
                  variant="ghost"
                  size="icon-sm"
                  render={
                    <Link
                      href={`/app/instituicoes/${item.id}`}
                      aria-label={`Editar ${item.nome}`}
                    />
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
                  aria-label={
                    item.situacao === "ativa" ? `Inativar ${item.nome}` : `Ativar ${item.nome}`
                  }
                  onClick={() => onAlterarSituacao(item)}
                >
                  {item.situacao === "ativa" ? "Inativar" : "Ativar"}
                </LoadingButton>
              </div>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
