"use client"

import { use, useState } from "react"
import { useRouter } from "next/navigation"
import { Trilha } from "@/components/layout/trilha"
import { useGuardaDeSaida } from "@/components/shared/hooks/use-guarda-de-saida"
import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import { ErrorState } from "@/components/shared/ui/error-state"
import { Skeleton } from "@/components/ui/skeleton"
import { IndicadorInepForm } from "@/features/indicador-inep/components/indicador-inep-form"
import { useIndicadorInep } from "@/features/indicador-inep/hooks/use-indicador-inep"

export default function EditarIndicadorInepPage({
  params,
}: PageProps<"/app/indicadores-inep/[id]">) {
  const { id } = use(params)
  const router = useRouter()
  const [sujo, setSujo] = useState(false)
  const { confirmando, confirmarDescarte, cancelarDescarte, solicitarSaida } = useGuardaDeSaida(sujo)
  const { data: indicador, isLoading, error, recarregar } = useIndicadorInep(id)

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Sistema" },
          { rotulo: "Indicadores do INEP", href: "/app/indicadores-inep" },
          { rotulo: indicador ? `${indicador.codigo} — ${indicador.nome}` : "Editar" },
        ]}
      />

      <h1 className="text-2xl font-semibold">Editar indicador do INEP</h1>

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
          titulo="Não foi possível carregar o indicador agora."
          onTentarNovamente={recarregar}
        />
      )}

      {!isLoading && !error && indicador && (
        <IndicadorInepForm
          key={indicador.versao}
          indicador={indicador}
          onDirtyChange={setSujo}
          onCancelar={() => solicitarSaida("/app/indicadores-inep")}
          onConflito={recarregar}
          onSalvo={() => router.push("/app/indicadores-inep")}
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
