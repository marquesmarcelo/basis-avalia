"use client"

import { useRouter, useSearchParams } from "next/navigation"
import { useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { useGuardaDeSaida } from "@/components/shared/hooks/use-guarda-de-saida"
import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import { MetaForm } from "@/features/meta/components/meta-form"

export default function NovaMetaPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const indicadorId = searchParams.get("indicador_id") ?? undefined
  const voltar = searchParams.get("voltar") || "/app/metas"

  const [sujo, setSujo] = useState(false)
  const { confirmando, confirmarDescarte, cancelarDescarte, solicitarSaida } =
    useGuardaDeSaida(sujo)

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Metas" },
          { rotulo: "Catálogo de metas", href: "/app/metas" },
          { rotulo: "Novo" },
        ]}
      />

      <h1 className="text-2xl font-semibold">Nova meta</h1>

      <MetaForm
        meta={null}
        indicadorPreSelecionadoId={indicadorId}
        criarOutro
        onDirtyChange={setSujo}
        onCancelar={() => solicitarSaida(voltar)}
        onSalvo={() => router.push(voltar)}
      />

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
