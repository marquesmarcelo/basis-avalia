"use client"

import { useRouter } from "next/navigation"
import { use, useEffect, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { useGuardaDeSaida } from "@/components/shared/hooks/use-guarda-de-saida"
import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import { ErrorState } from "@/components/shared/ui/error-state"
import { Skeleton } from "@/components/ui/skeleton"
import { CursoForm } from "@/features/curso/components/curso-form"
import { useCurso } from "@/features/curso/hooks/use-curso"

export default function EditarCursoPage({ params }: PageProps<"/app/cursos/[id]">) {
  const { id } = use(params)
  const router = useRouter()
  const [sujo, setSujo] = useState(false)
  const { confirmando, confirmarDescarte, cancelarDescarte, solicitarSaida } =
    useGuardaDeSaida(sujo)
  const { data: curso, isLoading, error, buscar } = useCurso()

  useEffect(() => {
    buscar(id)
  }, [id, buscar])

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Cursos", href: "/app/cursos" },
          { rotulo: curso?.nome ?? "Editar" },
        ]}
      />

      <h1 className="text-2xl font-semibold">Editar curso</h1>

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
          titulo="Não foi possível carregar o curso agora."
          onTentarNovamente={() => buscar(id)}
        />
      )}

      {!isLoading && !error && curso && (
        <CursoForm
          key={curso.versao}
          curso={curso}
          onDirtyChange={setSujo}
          onCancelar={() => solicitarSaida("/app/cursos")}
          onConflito={() => buscar(id)}
          onSalvo={() => router.push("/app/cursos")}
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
