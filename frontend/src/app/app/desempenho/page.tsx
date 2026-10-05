"use client"

import { BarChart3Icon } from "lucide-react"
import { useEffect, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { useEstadoDeGrid } from "@/components/shared/hooks/use-estado-de-grid"
import { EmptyState } from "@/components/shared/ui/empty-state"
import { ErrorState } from "@/components/shared/ui/error-state"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { Paginacao } from "@/components/shared/ui/paginacao"
import { SkeletonTable } from "@/components/shared/ui/skeleton-table"
import { useEu } from "@/features/auth/hooks/use-eu"
import { usePeriodosParaMinhasMetas } from "@/features/entrega/hooks/use-periodos-para-minhas-metas"
import { usePeriodosSugestoes } from "@/features/plano/hooks/use-periodos-sugestoes"
import { GraficoDesempenhoPorCurso } from "@/features/relatorio-desempenho/components/grafico-desempenho-por-curso"
import { RelatorioCardMobile } from "@/features/relatorio-desempenho/components/relatorio-card-mobile"
import { RelatorioFiltro } from "@/features/relatorio-desempenho/components/relatorio-filtro"
import { RelatorioTable } from "@/features/relatorio-desempenho/components/relatorio-table"
import { useDesempenhoPorCurso } from "@/features/relatorio-desempenho/hooks/use-desempenho-por-curso"
import { useExportarRelatorio } from "@/features/relatorio-desempenho/hooks/use-exportar-relatorio"
import { useRelatorioDesempenho } from "@/features/relatorio-desempenho/hooks/use-relatorio-desempenho"
import type { FiltroRelatorio } from "@/features/relatorio-desempenho/types"
import { possuiPermissao } from "@/lib/permissoes"

export default function DesempenhoPage() {
  const { data: eu } = useEu()
  const podeExportar = possuiPermissao(eu?.permissoes ?? [], "relatorio.exportar")
  const ehPI = possuiPermissao(eu?.permissoes ?? [], "relatorio.ler")

  const periodosPI = usePeriodosSugestoes()
  const periodosCoordenador = usePeriodosParaMinhasMetas()

  const {
    estado,
    carregado,
    definirFiltros,
    ordenarPor,
    definirPagina,
    definirTamanhoDePagina,
    ajustarPaginaAoTotal,
  } = useEstadoDeGrid<FiltroRelatorio>(
    "desempenho",
    {
      periodo_id: "",
      curso_id: "",
      meta_id: "",
      indicador_id: "",
      origem: "",
      situacao: "",
      autoavaliado: "",
      incluir_inativos: "",
    },
    "curso",
    "asc",
    20
  )
  const { data, isLoading, error, pesquisar } = useRelatorioDesempenho()
  const {
    data: dataPorCurso,
    isLoading: isLoadingPorCurso,
    error: errorPorCurso,
    pesquisar: pesquisarPorCurso,
  } = useDesempenhoPorCurso()
  const { exportar, isExporting } = useExportarRelatorio()
  const [jaPesquisou, setJaPesquisou] = useState(false)

  useEffect(() => {
    if (ehPI) periodosPI.buscar("")
    else periodosCoordenador.carregar()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ehPI])

  const opcoesPeriodo = ehPI
    ? periodosPI.itens.map((p) => ({ id: p.id, nome: p.nome }))
    : periodosCoordenador.itens.map((p) => ({ id: p.id, nome: p.nome }))

  async function executarPesquisaTabela() {
    if (!estado.filtros.periodo_id) return
    const resultado = await pesquisar(estado.filtros, {
      page: estado.page,
      page_size: estado.pageSize,
      sort: estado.sort,
      order: estado.order,
    })
    if (resultado) ajustarPaginaAoTotal(resultado.meta.total_pages)
  }

  async function executarPesquisa() {
    if (!estado.filtros.periodo_id) return
    setJaPesquisou(true)
    await Promise.all([executarPesquisaTabela(), pesquisarPorCurso(estado.filtros)])
  }

  // biome-ignore lint/correctness/useExhaustiveDependencies: dispara só por paginação/ordenação, nunca por jaPesquisou/carregado/executarPesquisaTabela
  useEffect(() => {
    if (carregado && jaPesquisou) executarPesquisaTabela()
  }, [estado.page, estado.pageSize, estado.sort, estado.order])

  function handlePesquisar(filtros: FiltroRelatorio) {
    definirFiltros(filtros)
    setJaPesquisou(true)
  }

  // biome-ignore lint/correctness/useExhaustiveDependencies: dispara só por mudança de filtro, nunca por jaPesquisou/executarPesquisa
  useEffect(() => {
    if (jaPesquisou) executarPesquisa()
  }, [estado.filtros])

  const itens = data?.data ?? []
  const tabelaVazia = jaPesquisou && !isLoading && !error && itens.length === 0

  return (
    <div className="w-full space-y-6">
      <Trilha itens={[{ rotulo: "Início", href: "/app" }, { rotulo: "Desempenho dos cursos" }]} />

      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Desempenho dos cursos</h1>
        {podeExportar && (
          <LoadingButton
            variant="outline"
            loading={isExporting}
            loadingText="Exportando..."
            disabled={!estado.filtros.periodo_id}
            onClick={() =>
              exportar(estado.filtros, {
                page: 1,
                page_size: estado.pageSize,
                sort: estado.sort,
                order: estado.order,
              })
            }
          >
            Exportar CSV
          </LoadingButton>
        )}
      </div>

      {data && data.resumo.cursos_sem_coordenador > 0 && (
        <p role="status" className="rounded-md border border-amber-400/50 bg-amber-50 p-3 text-sm">
          {data.resumo.cursos_sem_coordenador} curso(s) sem coordenador acumulam{" "}
          {data.resumo.metas_nao_cumpridas_de_vagos} meta(s) não cumprida(s) neste período.
        </p>
      )}

      <RelatorioFiltro
        filtrosIniciais={estado.filtros}
        opcoesPeriodo={opcoesPeriodo}
        onPesquisar={handlePesquisar}
        pesquisando={isLoading}
      />

      {jaPesquisou && !tabelaVazia && (
        <GraficoDesempenhoPorCurso
          itens={dataPorCurso?.data ?? []}
          cursosSemCoordenador={dataPorCurso?.cursos_sem_coordenador ?? []}
          meta={dataPorCurso?.meta ?? null}
          isLoading={isLoadingPorCurso}
          error={errorPorCurso}
          onTentarNovamente={() => pesquisarPorCurso(estado.filtros)}
        />
      )}

      {!jaPesquisou && (
        <EmptyState
          icon={BarChart3Icon}
          titulo="Selecione o período e clique em Pesquisar para ver o desempenho."
        />
      )}
      {jaPesquisou && isLoading && (
        <SkeletonTable
          colunas={[
            "Curso",
            "Responsável",
            "Meta",
            "Exigido",
            "Aceitas",
            "Cumprimento",
            "Situação",
          ]}
        />
      )}
      {jaPesquisou && !isLoading && error && (
        <ErrorState
          titulo="Não foi possível carregar o relatório agora."
          onTentarNovamente={executarPesquisaTabela}
        />
      )}
      {tabelaVazia && (
        <EmptyState
          icon={BarChart3Icon}
          titulo="Nenhum item encontrado para os filtros escolhidos."
        />
      )}
      {jaPesquisou && !isLoading && !error && itens.length > 0 && (
        <div className="space-y-4" aria-busy={isLoading}>
          <p className="sr-only" aria-live="polite">
            {data?.meta.total} itens encontrados.
          </p>
          <div className="hidden md:block">
            <RelatorioTable
              itens={itens}
              sort={estado.sort}
              order={estado.order}
              onOrdenar={ordenarPor}
            />
          </div>
          <div className="grid gap-3 md:hidden">
            {itens.map((item) => (
              <RelatorioCardMobile key={item.item_plano_id} item={item} />
            ))}
          </div>
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
    </div>
  )
}
