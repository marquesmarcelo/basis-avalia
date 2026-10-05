"use client"

import { Building2Icon } from "lucide-react"
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
import { AtivarInativarInstituicaoDialog } from "@/features/instituicao/components/ativar-inativar-instituicao-dialog"
import { InstituicaoCardMobile } from "@/features/instituicao/components/instituicao-card-mobile"
import { InstituicaoFiltro } from "@/features/instituicao/components/instituicao-filtro"
import { InstituicaoTable } from "@/features/instituicao/components/instituicao-table"
import { useAtivarInativarInstituicao } from "@/features/instituicao/hooks/use-ativar-inativar-instituicao"
import { useInstituicoes } from "@/features/instituicao/hooks/use-instituicoes"
import type { FiltroInstituicoes, Instituicao } from "@/features/instituicao/types"

export default function InstituicoesPage() {
  const {
    estado,
    carregado,
    definirFiltros,
    ordenarPor,
    definirPagina,
    definirTamanhoDePagina,
    ajustarPaginaAoTotal,
  } = useEstadoDeGrid<FiltroInstituicoes>(
    "instituicoes",
    { busca: "", situacao: "todas" },
    "nome",
    "asc",
    20
  )
  const { data, isLoading, error, pesquisar } = useInstituicoes()
  const { alterarSituacao, idEmAndamento } = useAtivarInativarInstituicao()

  const [jaPesquisou, setJaPesquisou] = useState(false)
  const [instituicaoParaInativar, setInstituicaoParaInativar] = useState<Instituicao | null>(null)

  async function executarPesquisa() {
    setJaPesquisou(true)
    const resultado = await pesquisar(estado.filtros, {
      page: estado.page,
      page_size: estado.pageSize,
      sort: estado.sort,
      order: estado.order,
    })
    if (resultado) {
      ajustarPaginaAoTotal(resultado.meta.total_pages)
    }
  }

  useEffect(() => {
    if (carregado && jaPesquisou) {
      executarPesquisa()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [estado.page, estado.pageSize, estado.sort, estado.order])

  function handlePesquisar(filtros: FiltroInstituicoes) {
    definirFiltros(filtros)
    setJaPesquisou(true)
  }

  useEffect(() => {
    if (jaPesquisou) {
      executarPesquisa()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [estado.filtros])

  function handleAlterarSituacao(item: Instituicao) {
    if (item.situacao === "ativa") {
      setInstituicaoParaInativar(item)
      return
    }
    ativar(item)
  }

  async function ativar(item: Instituicao) {
    const resultado = await alterarSituacao(item.id, "ativa", item.versao)
    if (resultado) {
      notificar.sucesso(`${item.nome} foi ativada.`)
      executarPesquisa()
    }
  }

  async function confirmarInativacao() {
    if (!instituicaoParaInativar) return
    const resultado = await alterarSituacao(
      instituicaoParaInativar.id,
      "inativa",
      instituicaoParaInativar.versao
    )
    if (resultado) {
      notificar.sucesso(`${instituicaoParaInativar.nome} foi inativada.`)
      setInstituicaoParaInativar(null)
      executarPesquisa()
    }
  }

  const itens = data?.data ?? []

  return (
    <div className="w-full space-y-6">
      <Trilha itens={[{ rotulo: "Início", href: "/app" }, { rotulo: "Instituições" }]} />

      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Instituições</h1>
        <LoadingButton id="botao-novo" render={<Link href="/app/instituicoes/novo" />}>
          + Nova
          <IndicadorDeNavegacao />
        </LoadingButton>
      </div>

      <InstituicaoFiltro
        filtrosIniciais={estado.filtros}
        onPesquisar={handlePesquisar}
        pesquisando={isLoading}
      />

      {!jaPesquisou && (
        <EmptyState
          icon={Building2Icon}
          titulo="Use os filtros acima e clique em Pesquisar para ver as instituições."
        />
      )}

      {jaPesquisou && isLoading && (
        <SkeletonTable
          colunas={["Nome", "Sigla", "Código e-MEC", "Situação", "Cadastrado em", "Ações"]}
        />
      )}

      {jaPesquisou && !isLoading && error && (
        <ErrorState
          titulo="Não foi possível carregar as instituições agora."
          onTentarNovamente={executarPesquisa}
        />
      )}

      {jaPesquisou && !isLoading && !error && itens.length === 0 && (
        <EmptyState
          icon={Building2Icon}
          titulo="Nenhuma instituição encontrada."
          descricao="Revise os filtros e tente novamente."
        />
      )}

      {jaPesquisou && !isLoading && !error && itens.length > 0 && (
        <div className="space-y-4" aria-busy={isLoading}>
          <p className="sr-only" aria-live="polite">
            {data?.meta.total} instituições encontradas.
          </p>
          <div className="hidden md:block">
            <InstituicaoTable
              itens={itens}
              sort={estado.sort}
              order={estado.order}
              onOrdenar={ordenarPor}
              onAlterarSituacao={handleAlterarSituacao}
              idAlterandoSituacao={idEmAndamento}
            />
          </div>
          <div className="grid gap-3 md:hidden">
            {itens.map((item) => (
              <InstituicaoCardMobile
                key={item.id}
                item={item}
                onAlterarSituacao={handleAlterarSituacao}
                idAlterandoSituacao={idEmAndamento}
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
        </div>
      )}

      <AtivarInativarInstituicaoDialog
        instituicao={instituicaoParaInativar}
        onOpenChange={(aberto) => !aberto && setInstituicaoParaInativar(null)}
        confirmando={idEmAndamento === instituicaoParaInativar?.id}
        onConfirmar={confirmarInativacao}
      />
    </div>
  )
}
