"use client"

import { FileTextIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { useEstadoDeGrid } from "@/components/shared/hooks/use-estado-de-grid"
import { EmptyState } from "@/components/shared/ui/empty-state"
import { ErrorState } from "@/components/shared/ui/error-state"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Paginacao } from "@/components/shared/ui/paginacao"
import { SkeletonTable } from "@/components/shared/ui/skeleton-table"
import { useEu } from "@/features/auth/hooks/use-eu"
import { ExcluirPlanoDialog } from "@/features/plano/components/excluir-plano-dialog"
import { PlanoCardMobile } from "@/features/plano/components/plano-card-mobile"
import { PlanoFiltro } from "@/features/plano/components/plano-filtro"
import { PlanoTable } from "@/features/plano/components/plano-table"
import { useExcluirPlano } from "@/features/plano/hooks/use-excluir-plano"
import { useGerarDocumento } from "@/features/plano/hooks/use-gerar-documento"
import { usePlanos } from "@/features/plano/hooks/use-planos"
import type { FiltroPlanos, Plano } from "@/features/plano/types"
import { possuiPermissao } from "@/lib/permissoes"

export default function PlanosPage() {
  const { data: eu } = useEu()
  const somenteLeitura = !possuiPermissao(eu?.permissoes ?? [], "plano.listar")
  const modo = somenteLeitura ? "meus-planos" : "planos"

  const {
    estado,
    carregado,
    definirFiltros,
    ordenarPor,
    definirPagina,
    definirTamanhoDePagina,
    ajustarPaginaAoTotal,
  } = useEstadoDeGrid<FiltroPlanos>(
    "planos",
    { periodo_id: "", curso_id: "", situacao: "todas", aprovacao: "todos" },
    "curso",
    "asc",
    20
  )
  const { data, isLoading, error, pesquisar } = usePlanos(modo)
  const { excluir, idEmAndamento: idExcluindo } = useExcluirPlano()
  const { gerar, isSubmitting: gerando } = useGerarDocumento(modo)

  const [jaPesquisou, setJaPesquisou] = useState(false)
  const [paraExcluir, setParaExcluir] = useState<Plano | null>(null)
  const [idGerando, setIdGerando] = useState<string | null>(null)

  async function executarPesquisa() {
    if (!eu) return
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
    if (carregado && jaPesquisou && eu) executarPesquisa()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [estado.page, estado.pageSize, estado.sort, estado.order, eu])

  function handlePesquisar(filtros: FiltroPlanos) {
    definirFiltros(filtros)
    setJaPesquisou(true)
  }

  useEffect(() => {
    if (jaPesquisou && eu) executarPesquisa()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [estado.filtros, eu])

  async function handleExcluir() {
    if (!paraExcluir) return
    const sucesso = await excluir(paraExcluir.id)
    if (sucesso) {
      notificar.sucesso("Plano excluído.")
      setParaExcluir(null)
      executarPesquisa()
    } else {
      notificar.erro("Este plano não pode ser excluído nesta situação.")
    }
  }

  async function handleGerarDocumento(item: Plano) {
    setIdGerando(item.id)
    const sucesso = await gerar(item.id)
    setIdGerando(null)
    if (sucesso) notificar.sucesso("Documento gerado e baixado.")
    else notificar.erro("Não foi possível gerar o documento agora. Tente novamente.")
  }

  const itens = data?.data ?? []

  return (
    <div className="w-full space-y-6">
      <Trilha itens={[{ rotulo: "Início", href: "/app" }, { rotulo: "Planos" }]} />

      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Planos de Ação Curso/Coordenador</h1>
        {!somenteLeitura && (
          <LoadingButton id="botao-novo" render={<a href="/app/planos/novo" />}>
            + Novo
          </LoadingButton>
        )}
      </div>

      <PlanoFiltro
        filtrosIniciais={estado.filtros}
        onPesquisar={handlePesquisar}
        pesquisando={isLoading}
        somenteLeitura={somenteLeitura}
      />

      {!jaPesquisou && (
        <EmptyState
          icon={FileTextIcon}
          titulo="Use os filtros acima e clique em Pesquisar para ver os planos."
        />
      )}

      {jaPesquisou && isLoading && (
        <SkeletonTable
          colunas={["Curso", "Período", "Situação", "Metas", "Exigido", "Aprovação", "Ações"]}
        />
      )}

      {jaPesquisou && !isLoading && error && (
        <ErrorState
          titulo="Não foi possível carregar os planos agora."
          onTentarNovamente={executarPesquisa}
        />
      )}

      {jaPesquisou && !isLoading && !error && itens.length === 0 && (
        <EmptyState icon={FileTextIcon} titulo="Nenhum plano encontrado." />
      )}

      {jaPesquisou && !isLoading && !error && itens.length > 0 && (
        <div className="space-y-4" aria-busy={isLoading}>
          <p className="sr-only" aria-live="polite">
            {data?.meta.total} planos encontrados.
          </p>
          <div className="hidden md:block">
            <PlanoTable
              itens={itens}
              sort={estado.sort}
              order={estado.order}
              onOrdenar={ordenarPor}
              somenteLeitura={somenteLeitura}
              onExcluir={setParaExcluir}
              onGerarDocumento={handleGerarDocumento}
              idExcluindo={idExcluindo}
              idGerandoDocumento={gerando ? idGerando : null}
            />
          </div>
          <div className="grid gap-3 md:hidden">
            {itens.map((item) => (
              <PlanoCardMobile
                key={item.id}
                item={item}
                somenteLeitura={somenteLeitura}
                onExcluir={setParaExcluir}
                onGerarDocumento={handleGerarDocumento}
                idExcluindo={idExcluindo}
                idGerandoDocumento={gerando ? idGerando : null}
              />
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
          <p className="text-sm text-muted-foreground">
            «Exigido» é a soma das quantidades dos itens — sempre por curso, nunca dividida entre
            cursos nem multiplicada pelo número de indicadores.
          </p>
        </div>
      )}

      {!somenteLeitura && (
        <ExcluirPlanoDialog
          plano={paraExcluir}
          confirmando={idExcluindo === paraExcluir?.id}
          onOpenChange={(aberto) => !aberto && setParaExcluir(null)}
          onConfirmar={handleExcluir}
        />
      )}
    </div>
  )
}
