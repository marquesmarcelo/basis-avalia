"use client"

import { useRouter } from "next/navigation"
import { use, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { useGuardaDeSaida } from "@/components/shared/hooks/use-guarda-de-saida"
import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import { ErrorState } from "@/components/shared/ui/error-state"
import { Skeleton } from "@/components/ui/skeleton"
import { useEu } from "@/features/auth/hooks/use-eu"
import { useInstituicao } from "@/features/instituicao/hooks/use-instituicao"
import { UsuarioForm } from "@/features/usuario/components/usuario-form"
import { useUsuario } from "@/features/usuario/hooks/use-usuario"

export default function EditarPesquisadorPage({
  params,
}: PageProps<"/app/instituicoes/[id]/pesquisadores/[usuarioId]">) {
  const { id, usuarioId } = use(params)
  const basePath = `/instituicoes/${id}/pesquisadores`
  const router = useRouter()
  const { data: eu } = useEu()
  const { data: instituicao } = useInstituicao(id)
  const [sujo, setSujo] = useState(false)
  const { confirmando, confirmarDescarte, cancelarDescarte, solicitarSaida } =
    useGuardaDeSaida(sujo)
  const { data: usuario, isLoading, error, recarregar } = useUsuario(basePath, usuarioId)

  const destino = `/app/instituicoes/${id}/pesquisadores`

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Instituições", href: "/app/instituicoes" },
          { rotulo: instituicao?.nome ?? "Pesquisadores", href: destino },
          { rotulo: usuario?.nome ?? "Editar" },
        ]}
      />

      <h1 className="text-2xl font-semibold">Editar Pesquisador Institucional</h1>

      {isLoading && (
        <div className="max-w-3xl space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
          </div>
          <Skeleton className="h-16 w-full" />
        </div>
      )}

      {!isLoading && error && (
        <ErrorState
          titulo="Não foi possível carregar o Pesquisador Institucional agora."
          onTentarNovamente={recarregar}
        />
      )}

      {!isLoading && !error && usuario && eu && (
        <UsuarioForm
          key={usuario.versao}
          basePath={basePath}
          usuario={usuario}
          usuarioAtualId={eu.id}
          perfilFixo="pesquisador_institucional"
          mensagemSucesso="Pesquisador Institucional atualizado com sucesso."
          onDirtyChange={setSujo}
          onCancelar={() => solicitarSaida(destino)}
          onConflito={recarregar}
          onSalvo={() => router.push(destino)}
        />
      )}

      <ConfirmDialog
        open={confirmando}
        onOpenChange={(aberto) => !aberto && cancelarDescarte()}
        titulo="Descartar alterações?"
        descricao="Você tem alterações não salvas. Deseja descartá-las?"
        rotuloConfirmar="Descartar alterações"
        rotuloCancelar="Continuar editando"
        destrutivo
        onConfirmar={confirmarDescarte}
      />
    </div>
  )
}
