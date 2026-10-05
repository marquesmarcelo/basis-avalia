"use client"

import { useParams, useRouter } from "next/navigation"
import { useEffect, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { AreaDeAnexos } from "@/components/shared/forms/area-de-anexos"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { SkeletonTable } from "@/components/shared/ui/skeleton-table"
import { Textarea } from "@/components/ui/textarea"
import { useAdicionarAnexo } from "@/features/entrega/hooks/use-adicionar-anexo"
import { useCorrigirEntrega } from "@/features/entrega/hooks/use-corrigir-entrega"
import { useEntrega } from "@/features/entrega/hooks/use-entrega"
import { useRemoverAnexo } from "@/features/entrega/hooks/use-remover-anexo"

export default function CorrigirEntregaPage() {
  const params = useParams<{ id: string }>()
  const router = useRouter()
  const { data: entrega, isLoading, carregar } = useEntrega()
  const { corrigir, isSubmitting } = useCorrigirEntrega()
  const { adicionar } = useAdicionarAnexo()
  const { remover, idEmAndamento: removendoId } = useRemoverAnexo()
  const [observacao, setObservacao] = useState("")

  useEffect(() => {
    carregar(params.id)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params.id])

  useEffect(() => {
    if (entrega) setObservacao(entrega.observacao)
  }, [entrega])

  async function handleAdicionar(arquivo: File): Promise<boolean> {
    const resultado = await adicionar(params.id, arquivo)
    if (resultado) {
      await carregar(params.id)
      return true
    }
    notificar.erro("Não foi possível anexar este arquivo.")
    return false
  }

  async function handleRemover(anexoId: string): Promise<boolean> {
    const sucesso = await remover(anexoId)
    if (sucesso) await carregar(params.id)
    return sucesso
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!entrega) return
    const resultado = await corrigir(entrega.id, observacao, entrega.versao)
    if (resultado) {
      notificar.sucesso("Entrega reenviada para avaliação.")
      router.push(`/app/itens/${entrega.item_plano_id}/entregas`)
    } else {
      notificar.erro("Não foi possível reenviar a entrega agora.")
    }
  }

  if (isLoading || !entrega) {
    return (
      <div className="w-full max-w-2xl space-y-6">
        <SkeletonTable colunas={["Correção"]} linhas={3} />
      </div>
    )
  }

  return (
    <div className="w-full max-w-2xl space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Minhas metas", href: "/app/minhas-metas" },
          { rotulo: "Corrigir entrega" },
        ]}
      />

      <h1 className="text-2xl font-semibold">Corrigir entrega</h1>

      {entrega.motivo && (
        <p role="status" className="rounded-md border border-amber-400/50 bg-amber-50 p-3 text-sm">
          Recusada: «{entrega.motivo}».{" "}
          {entrega.prazo_correcao_ate && (
            <>Corrija até {new Date(entrega.prazo_correcao_ate).toLocaleString("pt-BR")}. </>
          )}
          Recusa {entrega.rodadas} de 3.
        </p>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        <fieldset disabled={isSubmitting} className="space-y-2">
          <label htmlFor="observacao" className="text-sm font-medium">
            Observação (opcional)
          </label>
          <Textarea
            id="observacao"
            value={observacao}
            onChange={(e) => setObservacao(e.target.value)}
          />
        </fieldset>

        <AreaDeAnexos
          modo="porArquivo"
          anexosExistentes={entrega.anexos}
          onAdicionarArquivo={handleAdicionar}
          onRemoverAnexo={handleRemover}
          removendoAnexoId={removendoId}
          disabled={isSubmitting}
        />

        <LoadingButton type="submit" loading={isSubmitting} loadingText="Reenviando...">
          Reenviar
        </LoadingButton>
      </form>
    </div>
  )
}
