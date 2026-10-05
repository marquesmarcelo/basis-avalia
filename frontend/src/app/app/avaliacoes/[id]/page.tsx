"use client"

import { InfoIcon } from "lucide-react"
import { useParams } from "next/navigation"
import { useEffect, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { SkeletonTable } from "@/components/shared/ui/skeleton-table"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Textarea } from "@/components/ui/textarea"
import { useAvaliar } from "@/features/avaliacao/hooks/use-avaliar"
import { useDesfazerAceitacao } from "@/features/avaliacao/hooks/use-desfazer-aceitacao"
import { useEntrega } from "@/features/entrega/hooks/use-entrega"

export default function AvaliarEntregaPage() {
  const params = useParams<{ id: string }>()
  const { data: entrega, isLoading, carregar } = useEntrega()
  const { avaliar, isSubmitting: avaliando } = useAvaliar()
  const { desfazer, isSubmitting: desfazendo } = useDesfazerAceitacao()
  const [motivoRecusa, setMotivoRecusa] = useState("")
  const [mostrarRecusa, setMostrarRecusa] = useState(false)
  const [desfazerAberto, setDesfazerAberto] = useState(false)
  const [motivoDesfazer, setMotivoDesfazer] = useState("")

  useEffect(() => {
    carregar(params.id)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params.id])

  async function handleAceitar() {
    if (!entrega) return
    const resultado = await avaliar(entrega.id, "aceita", "", entrega.versao)
    if (resultado) {
      notificar.sucesso("Entrega aceita.")
      carregar(entrega.id)
    } else {
      notificar.erro("Não foi possível concluir a avaliação agora.")
    }
  }

  async function handleRecusar() {
    if (!entrega) return
    if (!motivoRecusa.trim()) {
      notificar.erro("Informe o motivo da recusa.")
      return
    }
    const resultado = await avaliar(entrega.id, "recusada", motivoRecusa, entrega.versao)
    if (resultado) {
      notificar.sucesso("Entrega recusada.")
      setMostrarRecusa(false)
      setMotivoRecusa("")
      carregar(entrega.id)
    } else {
      notificar.erro("Não foi possível concluir a avaliação agora.")
    }
  }

  async function handleDesfazer() {
    if (!entrega) return
    if (!motivoDesfazer.trim()) {
      notificar.erro("Informe o motivo.")
      return
    }
    const resultado = await desfazer(entrega.id, motivoDesfazer, entrega.versao)
    if (resultado) {
      notificar.sucesso("Aceitação desfeita — o cumprimento deste item diminuiu.")
      setDesfazerAberto(false)
      setMotivoDesfazer("")
      carregar(entrega.id)
    } else {
      notificar.erro("Não foi possível desfazer a aceitação agora.")
    }
  }

  if (isLoading || !entrega) {
    return (
      <div className="w-full max-w-2xl space-y-6">
        <SkeletonTable colunas={["Avaliação"]} linhas={4} />
      </div>
    )
  }

  return (
    <div className="w-full max-w-2xl space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Avaliação de entregas", href: "/app/avaliacoes" },
          { rotulo: "Avaliar" },
        ]}
      />

      <div>
        <h1 className="text-2xl font-semibold">{entrega.curso_nome}</h1>
        <p className="text-muted-foreground">{entrega.meta_nome}</p>
      </div>

      {entrega.coordenado_pelo_avaliador && entrega.situacao === "pendente_avaliacao" && (
        <p role="status" className="rounded-md border border-amber-400/50 bg-amber-50 p-3 text-sm">
          <InfoIcon className="mr-1 inline size-4" aria-hidden="true" />
          Você coordena este curso — a avaliação ficará marcada como avaliação pelo próprio
          coordenador. Você pode prosseguir normalmente.
        </p>
      )}

      <div className="space-y-2 rounded-lg border p-4">
        <p className="text-sm">
          <span className="font-medium">Enviada por:</span> {entrega.enviada_por_nome}
        </p>
        {entrega.observacao && (
          <p className="text-sm">
            <span className="font-medium">Observação:</span> {entrega.observacao}
          </p>
        )}
        <div>
          <p className="mb-1 text-sm font-medium">Anexos</p>
          <ul className="space-y-1">
            {entrega.anexos.map((a) => (
              <li key={a.id} className="text-sm">
                <a
                  href={`${process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:3001/api/v1"}/anexos/${a.id}/conteudo`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary hover:underline"
                >
                  {a.nome_original}
                </a>
              </li>
            ))}
          </ul>
        </div>
      </div>

      {entrega.situacao === "pendente_avaliacao" && (
        <div className="space-y-4">
          <div className="flex gap-3">
            <LoadingButton loading={avaliando} loadingText="Avaliando..." onClick={handleAceitar}>
              Aceitar
            </LoadingButton>
            <LoadingButton variant="outline" onClick={() => setMostrarRecusa((v) => !v)}>
              Recusar
            </LoadingButton>
          </div>
          {mostrarRecusa && (
            <div className="space-y-2">
              <label htmlFor="motivo-recusa" className="text-sm font-medium">
                Motivo da recusa
              </label>
              <Textarea
                id="motivo-recusa"
                value={motivoRecusa}
                onChange={(e) => setMotivoRecusa(e.target.value)}
              />
              <LoadingButton loading={avaliando} loadingText="Avaliando..." onClick={handleRecusar}>
                Confirmar recusa
              </LoadingButton>
            </div>
          )}
        </div>
      )}

      {entrega.situacao === "aceita" && (
        <LoadingButton variant="destructive" onClick={() => setDesfazerAberto(true)}>
          Desfazer aceitação
        </LoadingButton>
      )}

      <Dialog open={desfazerAberto} onOpenChange={setDesfazerAberto}>
        <DialogContent className="max-w-md" role="alertdialog">
          <DialogHeader>
            <DialogTitle>Desfazer aceitação</DialogTitle>
            <DialogDescription>
              O cumprimento deste item vai diminuir e o coordenador será notificado.
            </DialogDescription>
          </DialogHeader>
          <Textarea
            aria-label="Motivo do desfazimento"
            value={motivoDesfazer}
            onChange={(e) => setMotivoDesfazer(e.target.value)}
            placeholder="Motivo obrigatório"
          />
          <DialogFooter>
            <LoadingButton
              variant="outline"
              onClick={() => setDesfazerAberto(false)}
              disabled={desfazendo}
            >
              Continuar editando
            </LoadingButton>
            <LoadingButton
              variant="destructive"
              loading={desfazendo}
              loadingText="Desfazendo..."
              onClick={handleDesfazer}
            >
              Desfazer aceitação
            </LoadingButton>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
