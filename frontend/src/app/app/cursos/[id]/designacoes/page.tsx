"use client"

import { UsersIcon } from "lucide-react"
import Link from "next/link"
import { use, useEffect, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { useEstadoDeGrid } from "@/components/shared/hooks/use-estado-de-grid"
import { EmptyState } from "@/components/shared/ui/empty-state"
import { ErrorState } from "@/components/shared/ui/error-state"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Paginacao } from "@/components/shared/ui/paginacao"
import { SkeletonTable } from "@/components/shared/ui/skeleton-table"
import { useCurso } from "@/features/curso/hooks/use-curso"
import { DesignacaoFiltro } from "@/features/designacao/components/designacao-filtro"
import { DesignacaoTable } from "@/features/designacao/components/designacao-table"
import { EncerrarDesignacaoModal } from "@/features/designacao/components/encerrar-designacao-modal"
import { ExcluirDesignacaoDialog } from "@/features/designacao/components/excluir-designacao-dialog"
import { useDesignacoes } from "@/features/designacao/hooks/use-designacoes"
import { useExcluirDesignacao } from "@/features/designacao/hooks/use-excluir-designacao"
import type { Designacao, FiltroDesignacoes } from "@/features/designacao/types"

export default function DesignacoesDoCursoPage({
  params,
}: PageProps<"/app/cursos/[id]/designacoes">) {
  const { id } = use(params)

  const { data: curso, buscar: buscarCurso } = useCurso()
  const {
    estado,
    carregado,
    definirFiltros,
    ordenarPor,
    definirPagina,
    definirTamanhoDePagina,
    ajustarPaginaAoTotal,
  } = useEstadoDeGrid<FiltroDesignacoes>(
    `designacoes:${id}`,
    { situacao: "todas", coordenador_id: "" },
    "data_inicio",
    "desc",
    20
  )
  const { data, isLoading, error, pesquisar } = useDesignacoes(id)
  const { excluir, idEmAndamento: idExcluindo } = useExcluirDesignacao()

  const [jaPesquisou, setJaPesquisou] = useState(false)
  const [paraEncerrar, setParaEncerrar] = useState<Designacao | null>(null)
  const [paraExcluir, setParaExcluir] = useState<Designacao | null>(null)

  useEffect(() => {
    buscarCurso(id)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  async function executarPesquisa() {
    setJaPesquisou(true)
    const resultado = await pesquisar(estado.filtros, {
      page: estado.page,
      page_size: estado.pageSize,
      sort: estado.sort,
      order: estado.order,
    })
    if (resultado) ajustarPaginaAoTotal(resultado.meta.total_pages)
  }

  useEffect(() => {
    if (carregado && jaPesquisou) executarPesquisa()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [estado.page, estado.pageSize, estado.sort, estado.order])

  function handlePesquisar(filtros: FiltroDesignacoes) {
    definirFiltros(filtros)
    setJaPesquisou(true)
  }

  useEffect(() => {
    if (jaPesquisou) executarPesquisa()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [estado.filtros])

  async function confirmarExclusao() {
    if (!paraExcluir) return
    const sucesso = await excluir(paraExcluir.id)
    if (sucesso) {
      notificar.sucesso("Designação excluída.")
      setParaExcluir(null)
      executarPesquisa()
    }
  }

  const itens = data?.data ?? []

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Administração" },
          { rotulo: "Cursos", href: "/app/cursos" },
          { rotulo: curso?.nome ?? "Curso" },
          { rotulo: "Designações" },
        ]}
      />

      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Designações · {curso?.nome ?? ""}</h1>
        <LoadingButton
          id="botao-novo"
          render={<Link href={`/app/cursos/${id}/designacoes/novo`} />}
        >
          + Nova
          <IndicadorDeNavegacao />
        </LoadingButton>
      </div>

      <DesignacaoFiltro
        filtrosIniciais={estado.filtros}
        onPesquisar={handlePesquisar}
        pesquisando={isLoading}
      />

      {!jaPesquisou && (
        <EmptyState
          icon={UsersIcon}
          titulo="Use os filtros acima e clique em Pesquisar para ver as designações."
        />
      )}

      {jaPesquisou && isLoading && (
        <SkeletonTable
          colunas={["Coordenador", "Portaria", "Início", "Fim", "Situação", "Ações"]}
        />
      )}

      {jaPesquisou && !isLoading && error && (
        <ErrorState
          titulo="Não foi possível carregar as designações agora."
          onTentarNovamente={executarPesquisa}
        />
      )}

      {jaPesquisou && !isLoading && !error && itens.length === 0 && (
        <EmptyState
          icon={UsersIcon}
          titulo="Nenhuma designação encontrada."
          descricao="Revise os filtros e tente novamente."
        />
      )}

      {jaPesquisou && !isLoading && !error && itens.length > 0 && (
        <div className="space-y-4" aria-busy={isLoading}>
          <p className="sr-only" aria-live="polite">
            {data?.meta.total} designações encontradas.
          </p>
          <DesignacaoTable
            itens={itens}
            cursoId={id}
            sort={estado.sort}
            order={estado.order}
            onOrdenar={ordenarPor}
            onEncerrar={setParaEncerrar}
            onExcluir={setParaExcluir}
            idEmAndamento={idExcluindo}
          />
          {data && (
            <Paginacao
              page={data.meta.page}
              pageSize={data.meta.page_size}
              total={data.meta.total}
              totalPages={data.meta.total_pages}
              onPageChange={definirPagina}
              onPageSizeChange={definirTamanhoDePagina}
            />
          )}
        </div>
      )}

      <EncerrarDesignacaoModal
        open={!!paraEncerrar}
        onOpenChange={(aberto) => !aberto && setParaEncerrar(null)}
        designacao={paraEncerrar}
        cursoNome={curso?.nome ?? ""}
        onSucesso={() => executarPesquisa()}
      />

      <ExcluirDesignacaoDialog
        designacao={paraExcluir}
        onOpenChange={(aberto) => !aberto && setParaExcluir(null)}
        confirmando={idExcluindo === paraExcluir?.id}
        onConfirmar={confirmarExclusao}
      />
    </div>
  )
}
