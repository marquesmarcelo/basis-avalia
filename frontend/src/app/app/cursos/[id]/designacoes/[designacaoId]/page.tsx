"use client"

import { useRouter } from "next/navigation"
import { use, useEffect, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { useGuardaDeSaida } from "@/components/shared/hooks/use-guarda-de-saida"
import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import { ErrorState } from "@/components/shared/ui/error-state"
import { Skeleton } from "@/components/ui/skeleton"
import { useCurso } from "@/features/curso/hooks/use-curso"
import { DesignacaoForm } from "@/features/designacao/components/designacao-form"
import { useDesignacao } from "@/features/designacao/hooks/use-designacao"

export default function EditarDesignacaoPage({
  params,
}: PageProps<"/app/cursos/[id]/designacoes/[designacaoId]">) {
  const { id, designacaoId } = use(params)
  const router = useRouter()
  const { data: curso, buscar: buscarCurso } = useCurso()
  const { data: designacao, isLoading, error, buscar } = useDesignacao()
  const [sujo, setSujo] = useState(false)
  const { confirmando, confirmarDescarte, cancelarDescarte, solicitarSaida } =
    useGuardaDeSaida(sujo)

  const destino = `/app/cursos/${id}/designacoes`

  useEffect(() => {
    buscarCurso(id)
    buscar(designacaoId)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, designacaoId])

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Cursos", href: "/app/cursos" },
          { rotulo: curso?.nome ?? "Curso", href: `/app/cursos/${id}` },
          { rotulo: "Designações", href: destino },
          { rotulo: designacao?.coordenador.nome ?? "Editar" },
        ]}
      />

      <h1 className="text-2xl font-semibold">Editar designação · {curso?.nome ?? ""}</h1>

      {isLoading && (
        <div className="max-w-3xl space-y-4">
          <Skeleton className="h-16 w-full" />
          <Skeleton className="h-16 w-full" />
          <div className="grid gap-4 md:grid-cols-2">
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
          </div>
        </div>
      )}

      {!isLoading && error && (
        <ErrorState
          titulo="Não foi possível carregar a designação agora."
          onTentarNovamente={() => buscar(designacaoId)}
        />
      )}

      {!isLoading && !error && designacao && (
        <DesignacaoForm
          key={designacao.versao}
          cursoId={id}
          designacao={designacao}
          onDirtyChange={setSujo}
          onCancelar={() => solicitarSaida(destino)}
          onConflito={() => buscar(designacaoId)}
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
