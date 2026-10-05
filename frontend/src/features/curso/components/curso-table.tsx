import { BanIcon, PencilIcon, RotateCcwIcon, Trash2Icon, UsersIcon } from "lucide-react"
import Link from "next/link"
import { CabecalhoOrdenavel, type Direcao } from "@/components/shared/ui/cabecalho-ordenavel"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHeader, TableRow } from "@/components/ui/table"
import type { Curso } from "../types"
import { CursoCoordenadorCelula } from "./curso-coordenador-celula"

const rotuloGrau: Record<Curso["grau"], string> = {
  bacharelado: "Bacharelado",
  licenciatura: "Licenciatura",
  tecnologo: "Tecnólogo",
}

const rotuloModalidade: Record<Curso["modalidade"], string> = {
  presencial: "Presencial",
  a_distancia: "A distância",
}

interface CursoTableProps {
  itens: Curso[]
  sort: string
  order: Direcao
  onOrdenar: (campo: string) => void
  onVerDesignacoes: (item: Curso) => void
  onAlterarSituacao: (item: Curso) => void
  onExcluir: (item: Curso) => void
  idAlterandoSituacao: string | null
}

export function CursoTable({
  itens,
  sort,
  order,
  onOrdenar,
  onVerDesignacoes,
  onAlterarSituacao,
  onExcluir,
  idAlterandoSituacao,
}: CursoTableProps) {
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
            campo="codigo_emec"
            rotulo="Código e-MEC"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
            className="hidden md:table-cell"
          />
          <CabecalhoOrdenavel
            campo="grau"
            rotulo="Grau"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
          />
          <CabecalhoOrdenavel
            campo="modalidade"
            rotulo="Modalidade"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
          />
          <CabecalhoOrdenavel
            campo="coordenador"
            rotulo="Coordenador"
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
            campo="plano"
            rotulo="Plano do período corrente"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
            ordenavel={false}
            className="hidden md:table-cell"
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
            <TableCell className="hidden md:table-cell">{item.codigo_emec ?? "—"}</TableCell>
            <TableCell>{rotuloGrau[item.grau]}</TableCell>
            <TableCell>{rotuloModalidade[item.modalidade]}</TableCell>
            <TableCell>
              <CursoCoordenadorCelula coordenador={item.coordenador} />
            </TableCell>
            <TableCell>
              <StatusBadge
                label={item.situacao === "ativo" ? "Ativo" : "Inativo"}
                variant={item.situacao === "ativo" ? "default" : "secondary"}
              />
            </TableCell>
            <TableCell className="hidden md:table-cell">{item.plano_do_periodo ?? "—"}</TableCell>
            <TableCell>
              <div className="flex items-center gap-1">
                <Button
                  variant="ghost"
                  size="icon-sm"
                  render={
                    <Link href={`/app/cursos/${item.id}`} aria-label={`Editar ${item.nome}`} />
                  }
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
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
