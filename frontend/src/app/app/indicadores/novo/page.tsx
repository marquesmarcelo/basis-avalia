"use client"

import { useRouter } from "next/navigation"
import { useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { useGuardaDeSaida } from "@/components/shared/hooks/use-guarda-de-saida"
import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import { IndicadorForm } from "@/features/indicador/components/indicador-form"

export default function NovoIndicadorPage() {
  const router = useRouter()
  const [sujo, setSujo] = useState(false)
  const { confirmando, confirmarDescarte, cancelarDescarte, solicitarSaida } =
    useGuardaDeSaida(sujo)

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Metas" },
          { rotulo: "Indicadores", href: "/app/indicadores" },
          { rotulo: "Novo" },
        ]}
      />

      <h1 className="text-2xl font-semibold">Novo indicador</h1>

      <IndicadorForm
        indicador={null}
        criarOutro
        onDirtyChange={setSujo}
        onCancelar={() => solicitarSaida("/app/indicadores")}
        onSalvo={() => router.push("/app/indicadores")}
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
