"use client"

import { useRouter } from "next/navigation"
import { useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { useGuardaDeSaida } from "@/components/shared/hooks/use-guarda-de-saida"
import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import { InstituicaoForm } from "@/features/instituicao/components/instituicao-form"
import type { Instituicao } from "@/features/instituicao/types"

export default function NovaInstituicaoPage() {
  const router = useRouter()
  const [sujo, setSujo] = useState(false)
  const { confirmando, confirmarDescarte, cancelarDescarte, solicitarSaida } =
    useGuardaDeSaida(sujo)

  function handleSalvo(instituicao: Instituicao) {
    router.push(`/app/instituicoes/${instituicao.id}/pesquisadores/novo?primeiro=true`)
  }

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Instituições", href: "/app/instituicoes" },
          { rotulo: "Novo" },
        ]}
      />

      <h1 className="text-2xl font-semibold">Nova instituição</h1>

      <InstituicaoForm
        instituicao={null}
        criarOutro
        onDirtyChange={setSujo}
        onCancelar={() => solicitarSaida("/app/instituicoes")}
        onSalvo={handleSalvo}
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
