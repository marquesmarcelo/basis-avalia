"use client"

import { useParams, useRouter } from "next/navigation"
import { useCallback, useEffect, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import { ErrorState } from "@/components/shared/ui/error-state"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Skeleton } from "@/components/ui/skeleton"
import { useEu } from "@/features/auth/hooks/use-eu"
import { EncerrarDialog } from "@/features/plano/components/encerrar-dialog"
import { PlanoAvisoAprovacao } from "@/features/plano/components/plano-aviso-aprovacao"
import { PlanoCabecalho } from "@/features/plano/components/plano-cabecalho"
import { PlanoDadosForm } from "@/features/plano/components/plano-dados-form"
import { PlanoItensTabela } from "@/features/plano/components/plano-itens-tabela"
import { PublicarDialog } from "@/features/plano/components/publicar-dialog"
import { useDespublicarPlano } from "@/features/plano/hooks/use-despublicar-plano"
import { useEncerrarPlano } from "@/features/plano/hooks/use-encerrar-plano"
import { useGerarDocumento } from "@/features/plano/hooks/use-gerar-documento"
import { usePlano } from "@/features/plano/hooks/use-plano"
import { usePublicarPlano } from "@/features/plano/hooks/use-publicar-plano"
import { useReabrirPlano } from "@/features/plano/hooks/use-reabrir-plano"
import { possuiPermissao } from "@/lib/permissoes"

export default function PlanoDetalhePage() {
  const params = useParams<{ id: string }>()
  const router = useRouter()
  const { data: eu } = useEu()
  const somenteLeitura = !possuiPermissao(eu?.permissoes ?? [], "plano.gerenciar")
  const modo = possuiPermissao(eu?.permissoes ?? [], "plano.listar") ? "planos" : "meus-planos"

  const { data: plano, isLoading, error, carregar, definir } = usePlano(modo)
  const { gerar, isSubmitting: gerando } = useGerarDocumento(modo)
  const { publicar, isSubmitting: publicando } = usePublicarPlano()
  const { despublicar, isSubmitting: despublicando } = useDespublicarPlano()
  const { encerrar, isSubmitting: encerrando } = useEncerrarPlano()
  const { reabrir, isSubmitting: reabrindo } = useReabrirPlano()

  const [confirmarPublicar, setConfirmarPublicar] = useState(false)
  const [confirmarDespublicar, setConfirmarDespublicar] = useState(false)
  const [confirmarEncerrar, setConfirmarEncerrar] = useState(false)

  const carregarPlano = useCallback(() => {
    if (params.id) carregar(params.id)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params.id, modo])

  useEffect(() => {
    if (eu) carregarPlano()
  }, [eu, carregarPlano])

  async function handleGerarDocumento() {
    const sucesso = await gerar(params.id)
    if (sucesso) notificar.sucesso("Documento gerado e baixado.")
    else notificar.erro("Não foi possível gerar o documento agora. Tente novamente.")
  }

  async function handlePublicar() {
    if (!plano) return
    const resultado = await publicar(plano.id, plano.versao)
    if (resultado) {
      notificar.sucesso("Plano publicado.")
      definir(resultado)
      setConfirmarPublicar(false)
    } else {
      notificar.erro("Não foi possível publicar o plano agora.")
    }
  }

  async function handleDespublicar() {
    if (!plano) return
    const resultado = await despublicar(plano.id, plano.versao)
    if (resultado) {
      notificar.sucesso("Plano despublicado.")
      definir(resultado)
      setConfirmarDespublicar(false)
    } else {
      notificar.erro("Não foi possível despublicar o plano agora.")
    }
  }

  async function handleEncerrar(motivo: string) {
    if (!plano) return
    const resultado = await encerrar(plano.id, motivo, plano.versao)
    if (resultado) {
      notificar.sucesso("Plano encerrado.")
      definir(resultado)
      setConfirmarEncerrar(false)
    } else {
      notificar.erro("Não foi possível encerrar o plano agora.")
    }
  }

  async function handleReabrir() {
    if (!plano) return
    const versaoAnterior = plano.versao
    const resultado = await reabrir(plano.id, plano.versao)
    if (resultado) {
      definir(resultado)
      if (resultado.situacao === "encerrado" && resultado.versao > versaoAnterior) {
        notificar.info("O período já encerrou; o plano continua encerrado.")
      } else {
        notificar.sucesso("Plano reaberto.")
      }
    } else {
      notificar.erro("Não foi possível reabrir o plano agora.")
    }
  }

  if (isLoading || !plano) {
    return (
      <div className="w-full space-y-6">
        <Trilha
          itens={[
            { rotulo: "Início", href: "/app" },
            { rotulo: "Planos", href: "/app/planos" },
            { rotulo: "..." },
          ]}
        />
        {error ? (
          <ErrorState
            titulo="Este plano não existe ou você não tem acesso a ele."
            onTentarNovamente={() => router.push("/app/planos")}
            rotuloAcao="Voltar para Planos"
          />
        ) : (
          <div className="space-y-4">
            <Skeleton className="h-24 w-full" />
            <Skeleton className="h-48 w-full" />
            <Skeleton className="h-64 w-full" />
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Planos", href: "/app/planos" },
          { rotulo: `${plano.curso.nome} — ${plano.periodo.nome}` },
        ]}
      />

      <PlanoCabecalho
        plano={plano}
        somenteLeitura={somenteLeitura}
        gerandoDocumento={gerando}
        publicando={publicando}
        despublicando={despublicando}
        reabrindo={reabrindo}
        onGerarDocumento={handleGerarDocumento}
        onPublicar={() => setConfirmarPublicar(true)}
        onDespublicar={() => setConfirmarDespublicar(true)}
        onEncerrar={() => setConfirmarEncerrar(true)}
        onReabrir={handleReabrir}
      />

      <PlanoAvisoAprovacao plano={plano} />

      <section className="rounded-lg border p-4">
        <h2 className="mb-4 text-lg font-medium">Dados do plano</h2>
        <PlanoDadosForm
          plano={plano}
          somenteLeitura={somenteLeitura || plano.situacao === "encerrado"}
          onSalvo={definir}
          onConflito={carregarPlano}
        />
      </section>

      <section className="rounded-lg border p-4">
        <PlanoItensTabela
          plano={plano}
          somenteLeitura={somenteLeitura || plano.situacao === "encerrado"}
          onAtualizado={definir}
        />
      </section>

      <PublicarDialog
        plano={confirmarPublicar ? plano : null}
        onOpenChange={(open) => !open && setConfirmarPublicar(false)}
        confirmando={publicando}
        onConfirmar={handlePublicar}
      />

      <ConfirmDialog
        open={confirmarDespublicar}
        onOpenChange={setConfirmarDespublicar}
        titulo="Despublicar plano?"
        descricao={`Despublicar o plano de ${plano.curso.nome}? O coordenador deixará de vê-lo em Minhas metas, e as entregas deixam de ser possíveis.`}
        rotuloConfirmar="Despublicar"
        rotuloConfirmando="Despublicando..."
        confirmando={despublicando}
        onConfirmar={handleDespublicar}
      />

      <EncerrarDialog
        plano={confirmarEncerrar ? plano : null}
        onOpenChange={(open) => !open && setConfirmarEncerrar(false)}
        confirmando={encerrando}
        onConfirmar={handleEncerrar}
      />
    </div>
  )
}
