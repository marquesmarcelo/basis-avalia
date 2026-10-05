import { KeyIcon, PencilIcon, Trash2Icon } from "lucide-react"
import Link from "next/link"
import { CabecalhoOrdenavel, type Direcao } from "@/components/shared/ui/cabecalho-ordenavel"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHeader, TableRow } from "@/components/ui/table"
import { formatarData } from "@/lib/formato"
import type { Usuario } from "../types"

interface UsuarioTableProps {
  itens: Usuario[]
  sort: string
  order: Direcao
  onOrdenar: (campo: string) => void
  usuarioAtualId: string
  basePath: string
  onRedefinirSenha: (item: Usuario) => void
  onExcluir: (item: Usuario) => void
  idExcluindo: string | null
}

export function UsuarioTable({
  itens,
  sort,
  order,
  onOrdenar,
  usuarioAtualId,
  basePath,
  onRedefinirSenha,
  onExcluir,
  idExcluindo,
}: UsuarioTableProps) {
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
            campo="email"
            rotulo="E-mail"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
            className="hidden md:table-cell"
          />
          <CabecalhoOrdenavel
            campo="perfis"
            rotulo="Perfis"
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
        {itens.map((item) => {
          const propriaLinha = item.id === usuarioAtualId
          return (
            <TableRow key={item.id}>
              <TableCell>{item.nome}</TableCell>
              <TableCell className="hidden md:table-cell">{item.email}</TableCell>
              <TableCell>
                <span aria-hidden="true">
                  {item.perfis_rotulos.slice(0, 2).join(", ")}
                  {item.perfis_rotulos.length > 2 ? ` +${item.perfis_rotulos.length - 2}` : ""}
                </span>
                <span className="sr-only">{item.perfis_rotulos.join(", ")}</span>
              </TableCell>
              <TableCell>
                {item.senha_provisoria ? (
                  <StatusBadge label="Primeiro acesso pendente" variant="secondary" />
                ) : (
                  <StatusBadge label="Ativo" variant="default" />
                )}
              </TableCell>
              <TableCell className="hidden lg:table-cell">{formatarData(item.criado_em)}</TableCell>
              <TableCell>
                <div className="flex items-center gap-1">
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    render={
                      <Link
                        href={`/app${basePath}/${item.id}`}
                        aria-label={`Editar ${item.nome}`}
                      />
                    }
                  >
                    <PencilIcon aria-hidden="true" />
                    <IndicadorDeNavegacao />
                  </Button>
                  {!propriaLinha && (
                    <>
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        aria-label={`Redefinir senha de ${item.nome}`}
                        onClick={() => onRedefinirSenha(item)}
                      >
                        <KeyIcon aria-hidden="true" />
                      </Button>
                      <LoadingButton
                        variant="ghost"
                        size="icon-sm"
                        loading={idExcluindo === item.id}
                        aria-label={`Excluir ${item.nome}`}
                        onClick={() => onExcluir(item)}
                      >
                        <Trash2Icon aria-hidden="true" />
                      </LoadingButton>
                    </>
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
