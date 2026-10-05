"use client"

import { TargetIcon } from "lucide-react"
import Link from "next/link"
import { useEffect, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { useEstadoDeGrid } from "@/components/shared/hooks/use-estado-de-grid"
import { EmptyState } from "@/components/shared/ui/empty-state"
import { ErrorState } from "@/components/shared/ui/error-state"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Paginacao } from "@/components/shared/ui/paginacao"
import { SkeletonTable } from "@/components/shared/ui/skeleton-table"
import { ExcluirIndicadorInepDialog } from "@/features/indicador-inep/components/excluir-indicador-inep-dialog"
import { InativarIndicadorInepDialog } from "@/features/indicador-inep/components/inativar-indicador-inep-dialog"
import { IndicadorInepCardMobile } from "@/features/indicador-inep/components/indicador-inep-card-mobile"
import { IndicadorInepFiltro } from "@/features/indicador-inep/components/indicador-inep-filtro"
import { IndicadorInepTable } from "@/features/indicador-inep/components/indicador-inep-table"
import { useAlterarSituacaoIndicadorInep } from "@/features/indicador-inep/hooks/use-alterar-situacao-indicador-inep"
import { useExcluirIndicadorInep } from "@/features/indicador-inep/hooks/use-excluir-indicador-inep"
import { useIndicadoresInep } from "@/features/indicador-inep/hooks/use-indicadores-inep"
import type { FiltroIndicadoresInep, IndicadorInep } from "@/features/indicador-inep/types"

export default function IndicadoresInepPage() {
  const {
    estado,
    carregado,
    definirFiltros,
    ordenarPor,
    definirPagina,
    definirTamanhoDePagina,
    ajustarPaginaAoTotal,
  } = useEstadoDeGrid<FiltroIndicadoresInep>(
    "indicadores-inep",
    { busca: "", situacao: "ativo" },
    "codigo",
    "asc",
    20
  )
  const { data, isLoading, error, pesquisar } = useIndicadoresInep()
  const { alterarSituacao, idEmAndamento } = useAlterarSituacaoIndicadorInep()
  const { excluir, idEmAndamento: idExcluindo } = useExcluirIndicadorInep()

  const [jaPesquisou, setJaPesquisou] = useState(false)
  const [paraInativar, setParaInativar] = useState<IndicadorInep | null>(null)
  const [paraExcluir, setParaExcluir] = useState<IndicadorInep | null>(null)

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

  function handlePesquisar(filtros: FiltroIndicadoresInep) {
    definirFiltros(filtros)
    setJaPesquisou(true)
  }

  useEffect(() => {
    if (jaPesquisou) executarPesquisa()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [estado.filtros])

  function handleAlterarSituacao(item: IndicadorInep) {
    if (item.situacao === "ativo") {
      setParaInativar(item)
      return
    }
    reativar(item)
  }

  async function reativar(item: IndicadorInep) {
    const resultado = await alterarSituacao(item.id, "ativo", item.versao)
    if (resultado) {
      notificar.sucesso(`${item.codigo} foi reativado.`)
      executarPesquisa()
    }
  }

  async function confirmarInativacao() {
    if (!paraInativar) return
    const resultado = await alterarSituacao(paraInativar.id, "inativo", paraInativar.versao)
    if (resultado) {
      notificar.sucesso(`${paraInativar.codigo} foi inativado.`)
      setParaInativar(null)
      executarPesquisa()
    }
  }

  async function confirmarExclusao() {
    if (!paraExcluir) return
    const sucesso = await excluir(paraExcluir.id)
    if (sucesso) {
      notificar.sucesso(`${paraExcluir.codigo} foi excluído.`)
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
          { rotulo: "Sistema" },
          { rotulo: "Indicadores do INEP" },
        ]}
      />

      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Indicadores do INEP</h1>
          <p className="text-sm text-muted-foreground">
            Catálogo único da instalação. Todas as instituições leem; só você edita.
          </p>
        </div>
        <LoadingButton id="botao-novo" render={<Link href="/app/indicadores-inep/novo" />}>
          + Novo
          <IndicadorDeNavegacao />
        </LoadingButton>
      </div>

      <IndicadorInepFiltro
        filtrosIniciais={estado.filtros}
        onPesquisar={handlePesquisar}
        pesquisando={isLoading}
      />

      {!jaPesquisou && (
        <EmptyState
          icon={TargetIcon}
          titulo="Use os filtros acima e clique em Pesquisar para ver os indicadores."
        />
      )}

      {jaPesquisou && isLoading && (
        <SkeletonTable colunas={["Código", "Nome", "Referência", "Situação", "Metas", "Ações"]} />
      )}

      {jaPesquisou && !isLoading && error && (
        <ErrorState
          titulo="Não foi possível carregar os indicadores agora."
          onTentarNovamente={executarPesquisa}
        />
      )}

      {jaPesquisou && !isLoading && !error && itens.length === 0 && (
        <EmptyState
          icon={TargetIcon}
          titulo="Nenhum indicador encontrado."
          descricao="Revise os filtros e tente novamente."
        />
      )}

      {jaPesquisou && !isLoading && !error && itens.length > 0 && (
        <div className="space-y-4" aria-busy={isLoading}>
          <p className="sr-only" aria-live="polite">
            {data?.meta.total} indicadores encontrados.
          </p>
          <div className="hidden md:block">
            <IndicadorInepTable
              itens={itens}
              sort={estado.sort}
              order={estado.order}
              onOrdenar={ordenarPor}
              onAlterarSituacao={handleAlterarSituacao}
              onExcluir={setParaExcluir}
              idAlterandoSituacao={idEmAndamento ?? idExcluindo}
            />
          </div>
          <div className="grid gap-3 md:hidden">
            {itens.map((item) => (
              <IndicadorInepCardMobile
                key={item.id}
                item={item}
                onAlterarSituacao={handleAlterarSituacao}
                onExcluir={setParaExcluir}
                idAlterandoSituacao={idEmAndamento ?? idExcluindo}
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
            &quot;Metas&quot; é o total na instalação, somadas todas as instituições. Quais
            instituições usam não é informação desta tela.
          </p>
        </div>
      )}

      <InativarIndicadorInepDialog
        indicador={paraInativar}
        onOpenChange={(aberto) => !aberto && setParaInativar(null)}
        confirmando={idEmAndamento === paraInativar?.id}
        onConfirmar={confirmarInativacao}
      />

      <ExcluirIndicadorInepDialog
        indicador={paraExcluir}
        onOpenChange={(aberto) => !aberto && setParaExcluir(null)}
        confirmando={idExcluindo === paraExcluir?.id}
        onConfirmar={confirmarExclusao}
      />
    </div>
  )
}
