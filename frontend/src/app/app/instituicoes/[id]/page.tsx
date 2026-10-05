"use client"

import { useRouter } from "next/navigation"
import { use, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { useGuardaDeSaida } from "@/components/shared/hooks/use-guarda-de-saida"
import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import { ErrorState } from "@/components/shared/ui/error-state"
import { Skeleton } from "@/components/ui/skeleton"
import { InstituicaoForm } from "@/features/instituicao/components/instituicao-form"
import { useInstituicao } from "@/features/instituicao/hooks/use-instituicao"

export default function EditarInstituicaoPage({ params }: PageProps<"/app/instituicoes/[id]">) {
  const { id } = use(params)
  const router = useRouter()
  const [sujo, setSujo] = useState(false)
  const { confirmando, confirmarDescarte, cancelarDescarte, solicitarSaida } =
    useGuardaDeSaida(sujo)
  const { data: instituicao, isLoading, error, recarregar } = useInstituicao(id)

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Instituições", href: "/app/instituicoes" },
          { rotulo: instituicao?.nome ?? "Editar" },
        ]}
      />

      <h1 className="text-2xl font-semibold">Editar instituição</h1>

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
          titulo="Não foi possível carregar a instituição agora."
          onTentarNovamente={recarregar}
        />
      )}

      {!isLoading && !error && instituicao && (
        <InstituicaoForm
          key={instituicao.versao}
          instituicao={instituicao}
          onDirtyChange={setSujo}
          onCancelar={() => solicitarSaida("/app/instituicoes")}
          onConflito={recarregar}
          onSalvo={() => router.push("/app/instituicoes")}
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
