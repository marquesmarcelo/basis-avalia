"use client"

import { useRouter } from "next/navigation"
import { use, useEffect, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { useGuardaDeSaida } from "@/components/shared/hooks/use-guarda-de-saida"
import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import { useCurso } from "@/features/curso/hooks/use-curso"
import { DesignacaoForm } from "@/features/designacao/components/designacao-form"

export default function NovaDesignacaoPage({
  params,
}: PageProps<"/app/cursos/[id]/designacoes/novo">) {
  const { id } = use(params)
  const router = useRouter()
  const { data: curso, buscar: buscarCurso } = useCurso()
  const [sujo, setSujo] = useState(false)
  const { confirmando, confirmarDescarte, cancelarDescarte, solicitarSaida } =
    useGuardaDeSaida(sujo)

  const destino = `/app/cursos/${id}/designacoes`

  useEffect(() => {
    buscarCurso(id)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Cursos", href: "/app/cursos" },
          { rotulo: curso?.nome ?? "Curso", href: `/app/cursos/${id}` },
          { rotulo: "Designações", href: destino },
          { rotulo: "Nova" },
        ]}
      />

      <h1 className="text-2xl font-semibold">Nova designação · {curso?.nome ?? ""}</h1>

      <DesignacaoForm
        cursoId={id}
        designacao={null}
        criarOutro
        onDirtyChange={setSujo}
        onCancelar={() => solicitarSaida(destino)}
        onSalvo={() => router.push(destino)}
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
