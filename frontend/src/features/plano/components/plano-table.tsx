import { CopyIcon, DownloadIcon, Trash2Icon } from "lucide-react"
import Link from "next/link"
import { CabecalhoOrdenavel, type Direcao } from "@/components/shared/ui/cabecalho-ordenavel"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHeader, TableRow } from "@/components/ui/table"
import { formatarDataPura } from "@/lib/formato"
import type { Plano } from "../types"

const ROTULO_SITUACAO: Record<Plano["situacao"], string> = {
  rascunho: "Rascunho",
  vigente: "Vigente",
  encerrado: "Encerrado",
}

const VARIANTE_SITUACAO: Record<Plano["situacao"], "default" | "secondary"> = {
  rascunho: "secondary",
  vigente: "default",
  encerrado: "secondary",
}

interface PlanoTableProps {
  itens: Plano[]
  sort: string
  order: Direcao
  onOrdenar: (campo: string) => void
  somenteLeitura?: boolean
  onExcluir: (item: Plano) => void
  onGerarDocumento: (item: Plano) => void
  idExcluindo: string | null
  idGerandoDocumento: string | null
}

export function PlanoTable({
  itens,
  sort,
  order,
  onOrdenar,
  somenteLeitura,
  onExcluir,
  onGerarDocumento,
  idExcluindo,
  idGerandoDocumento,
}: PlanoTableProps) {
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
            campo="periodo"
            rotulo="Período"
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
            campo="exigido"
            rotulo="Exigido"
            sortAtivo={sort}
            orderAtivo={order}
            onOrdenar={onOrdenar}
            ordenavel={false}
          />
          <CabecalhoOrdenavel
            campo="aprovacao"
            rotulo="Aprovação"
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
            <TableCell>
              <Button
                variant="link"
                className="h-auto p-0"
                render={<Link href={`/app/planos/${item.id}`} />}
              >
                {item.curso.nome}
              </Button>
              {item.curso.vago && (
                <span className="ml-1 text-sm text-muted-foreground">Vago ⚠</span>
              )}
            </TableCell>
            <TableCell>{item.periodo.nome}</TableCell>
            <TableCell>
              <StatusBadge
                label={ROTULO_SITUACAO[item.situacao]}
                variant={VARIANTE_SITUACAO[item.situacao]}
              />
            </TableCell>
            <TableCell className="text-center">{item.metas}</TableCell>
            <TableCell className="text-center">{item.total_exigido}</TableCell>
            <TableCell>
              {item.aprovacao
                ? `${formatarDataPura(item.aprovacao.data)} · ${item.aprovacao.orgao === "nde" ? "NDE" : "Colegiado"}`
                : item.sem_aprovacao
                  ? "⚠ sem aprovação"
                  : "—"}
            </TableCell>
            <TableCell>
              <div className="flex items-center gap-1">
                {!somenteLeitura && (
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    aria-label={`Copiar plano de ${item.curso.nome}`}
                    render={<Link href={`/app/planos/${item.id}/copiar`} />}
                  >
                    <CopyIcon aria-hidden="true" />
                  </Button>
                )}
                <LoadingButton
                  variant="ghost"
                  size="icon-sm"
                  loading={idGerandoDocumento === item.id}
                  loadingText=""
                  aria-label={`Baixar documento de ${item.curso.nome}`}
                  onClick={() => onGerarDocumento(item)}
                >
                  <DownloadIcon aria-hidden="true" />
                </LoadingButton>
                {!somenteLeitura && item.situacao === "rascunho" && (
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    aria-label={`Excluir plano de ${item.curso.nome}`}
                    disabled={idExcluindo === item.id}
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
