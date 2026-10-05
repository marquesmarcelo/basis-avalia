"use client"

import { useRouter } from "next/navigation"
import { use, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { useGuardaDeSaida } from "@/components/shared/hooks/use-guarda-de-saida"
import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import { ErrorState } from "@/components/shared/ui/error-state"
import { Skeleton } from "@/components/ui/skeleton"
import { PeriodoForm } from "@/features/periodo/components/periodo-form"
import { usePeriodo } from "@/features/periodo/hooks/use-periodo"

export default function EditarPeriodoPage({ params }: PageProps<"/app/periodos/[id]">) {
  const { id } = use(params)
  const router = useRouter()
  const [sujo, setSujo] = useState(false)
  const { confirmando, confirmarDescarte, cancelarDescarte, solicitarSaida } =
    useGuardaDeSaida(sujo)
  const { data: periodo, isLoading, error, recarregar } = usePeriodo(id)

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Períodos", href: "/app/periodos" },
          { rotulo: periodo?.nome ?? "Editar" },
        ]}
      />

      <h1 className="text-2xl font-semibold">Editar período</h1>

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
          titulo="Não foi possível carregar o período agora."
          onTentarNovamente={recarregar}
        />
      )}

      {!isLoading && !error && periodo && (
        <PeriodoForm
          periodo={periodo}
          onDirtyChange={setSujo}
          onCancelar={() => solicitarSaida("/app/periodos")}
          onConflito={recarregar}
          onSalvo={() => router.push("/app/periodos")}
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
