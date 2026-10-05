"use client"

import { FileTextIcon } from "lucide-react"
import Link from "next/link"
import { useParams } from "next/navigation"
import { useEffect } from "react"
import { Trilha } from "@/components/layout/trilha"
import { EmptyState } from "@/components/shared/ui/empty-state"
import { ErrorState } from "@/components/shared/ui/error-state"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { SkeletonTable } from "@/components/shared/ui/skeleton-table"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { useEntregasDoItem } from "@/features/entrega/hooks/use-entregas-do-item"

function situacaoParaBadge(situacao: string): {
  label: string
  variant: "default" | "secondary" | "destructive"
} {
  switch (situacao) {
    case "aceita":
      return { label: "Aceita", variant: "default" }
    case "recusada":
      return { label: "Em correção", variant: "destructive" }
    default:
      return { label: "Pendente de avaliação", variant: "secondary" }
  }
}

export default function EntregasDoItemPage() {
  const params = useParams<{ itemId: string }>()
  const { data, isLoading, error, carregar } = useEntregasDoItem()

  useEffect(() => {
    carregar(params.itemId)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params.itemId])

  const itens = data?.data ?? []

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Minhas metas", href: "/app/minhas-metas" },
          { rotulo: "Entregas" },
        ]}
      />

      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Entregas deste item</h1>
        <LoadingButton render={<Link href={`/app/itens/${params.itemId}/entregas/nova`} />}>
          + Prestar contas
        </LoadingButton>
      </div>

      {isLoading && <SkeletonTable colunas={["Enviada por", "Data", "Situação"]} />}
      {!isLoading && error && (
        <ErrorState
          titulo="Não foi possível carregar as entregas agora."
          onTentarNovamente={() => carregar(params.itemId)}
        />
      )}
      {!isLoading && !error && itens.length === 0 && (
        <EmptyState icon={FileTextIcon} titulo="Nenhuma entrega registrada ainda." />
      )}
      {!isLoading && !error && itens.length > 0 && (
        <ul className="space-y-2">
          {itens.map((e) => {
            const badge = situacaoParaBadge(e.situacao)
            return (
              <li key={e.id} className="flex items-center justify-between rounded-lg border p-4">
                <div>
                  <p className="font-medium">{e.enviada_por_nome}</p>
                  <p className="text-sm text-muted-foreground">
                    {new Date(e.criado_em).toLocaleString("pt-BR")}
                  </p>
                </div>
                <div className="flex items-center gap-3">
                  <StatusBadge label={badge.label} variant={badge.variant} />
                  {e.situacao === "recusada" && (
                    <Link
                      href={`/app/entregas/${e.id}/corrigir`}
                      className="text-sm font-medium text-primary hover:underline"
                    >
                      Corrigir
                    </Link>
                  )}
                </div>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
