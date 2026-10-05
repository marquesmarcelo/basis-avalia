"use client"

import { useRouter } from "next/navigation"
import { use, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { useGuardaDeSaida } from "@/components/shared/hooks/use-guarda-de-saida"
import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import { ErrorState } from "@/components/shared/ui/error-state"
import { Skeleton } from "@/components/ui/skeleton"
import { useEu } from "@/features/auth/hooks/use-eu"
import { UsuarioForm } from "@/features/usuario/components/usuario-form"
import { useUsuario } from "@/features/usuario/hooks/use-usuario"

const BASE_PATH = "/usuarios"

export default function EditarUsuarioPage({ params }: PageProps<"/app/usuarios/[id]">) {
  const { id } = use(params)
  const router = useRouter()
  const { data: eu } = useEu()
  const [sujo, setSujo] = useState(false)
  const { confirmando, confirmarDescarte, cancelarDescarte, solicitarSaida } =
    useGuardaDeSaida(sujo)
  const { data: usuario, isLoading, error, recarregar } = useUsuario(BASE_PATH, id)

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Usuários", href: "/app/usuarios" },
          { rotulo: usuario?.nome ?? "Editar" },
        ]}
      />

      <h1 className="text-2xl font-semibold">Editar usuário</h1>

      {isLoading && (
        <div className="max-w-3xl space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
          </div>
          <Skeleton className="h-24 w-full" />
        </div>
      )}

      {!isLoading && error && (
        <ErrorState
          titulo="Não foi possível carregar o usuário agora."
          onTentarNovamente={recarregar}
        />
      )}

      {!isLoading && !error && usuario && eu && (
        <UsuarioForm
          key={usuario.versao}
          basePath={BASE_PATH}
          usuario={usuario}
          usuarioAtualId={eu.id}
          onDirtyChange={setSujo}
          onCancelar={() => solicitarSaida("/app/usuarios")}
          onConflito={recarregar}
          onSalvo={() => router.push("/app/usuarios")}
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
