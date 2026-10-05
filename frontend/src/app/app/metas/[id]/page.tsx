"use client"

import { useRouter, useSearchParams } from "next/navigation"
import { use, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { useGuardaDeSaida } from "@/components/shared/hooks/use-guarda-de-saida"
import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import { ErrorState } from "@/components/shared/ui/error-state"
import { Skeleton } from "@/components/ui/skeleton"
import { MetaForm } from "@/features/meta/components/meta-form"
import { useMeta } from "@/features/meta/hooks/use-meta"

export default function EditarMetaPage({ params }: PageProps<"/app/metas/[id]">) {
  const { id } = use(params)
  const router = useRouter()
  const searchParams = useSearchParams()
  const voltar = searchParams.get("voltar") || "/app/metas"

  const [sujo, setSujo] = useState(false)
  const { confirmando, confirmarDescarte, cancelarDescarte, solicitarSaida } =
    useGuardaDeSaida(sujo)
  const { data: meta, isLoading, error, recarregar } = useMeta(id)

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Metas" },
          { rotulo: "Catálogo de metas", href: "/app/metas" },
          { rotulo: meta?.nome ?? "Editar" },
        ]}
      />

      <h1 className="text-2xl font-semibold">Editar meta</h1>

      {isLoading && (
        <div className="max-w-3xl space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
          </div>
          <Skeleton className="h-16 w-full" />
          <Skeleton className="h-24 w-full" />
        </div>
      )}

      {!isLoading && error && (
        <ErrorState
          titulo="Não foi possível carregar a meta agora."
          onTentarNovamente={recarregar}
        />
      )}

      {!isLoading && !error && meta && (
        <MetaForm
          key={meta.versao}
          meta={meta}
          onDirtyChange={setSujo}
          onCancelar={() => solicitarSaida(voltar)}
          onConflito={recarregar}
          onSalvo={() => router.push(voltar)}
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
