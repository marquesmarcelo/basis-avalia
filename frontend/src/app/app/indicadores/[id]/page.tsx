"use client"

import { ClipboardListIcon, PlusIcon } from "lucide-react"
import Link from "next/link"
import { use, useEffect, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { SelectComRotulo } from "@/components/shared/forms/select-com-rotulo"
import { useEstadoDeGrid } from "@/components/shared/hooks/use-estado-de-grid"
import { useGuardaDeSaida } from "@/components/shared/hooks/use-guarda-de-saida"
import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import { EmptyState } from "@/components/shared/ui/empty-state"
import { ErrorState } from "@/components/shared/ui/error-state"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Paginacao } from "@/components/shared/ui/paginacao"
import { SkeletonTable } from "@/components/shared/ui/skeleton-table"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Field, FieldContent, FieldLabel } from "@/components/ui/field"
import { Skeleton } from "@/components/ui/skeleton"
import { IndicadorForm } from "@/features/indicador/components/indicador-form"
import { useIndicador } from "@/features/indicador/hooks/use-indicador"
import { ExcluirMetaDialog } from "@/features/meta/components/excluir-meta-dialog"
import { InativarMetaDialog } from "@/features/meta/components/inativar-meta-dialog"
import { MetaCardMobile } from "@/features/meta/components/meta-card-mobile"
import { MetaTable } from "@/features/meta/components/meta-table"
import { useAlterarSituacaoMeta } from "@/features/meta/hooks/use-alterar-situacao-meta"
import { useAtualizarMeta } from "@/features/meta/hooks/use-atualizar-meta"
import { useExcluirMeta } from "@/features/meta/hooks/use-excluir-meta"
import { useMetas } from "@/features/meta/hooks/use-metas"
import type { FiltroMetas, Meta } from "@/features/meta/types"

export default function EditarIndicadorPage({ params }: PageProps<"/app/indicadores/[id]">) {
  const { id } = use(params)
  const [sujo, setSujo] = useState(false)
  const { confirmando, confirmarDescarte, cancelarDescarte, solicitarSaida } =
    useGuardaDeSaida(sujo)
  const { data: indicador, isLoading, error, recarregar } = useIndicador(id)

  const proprio = indicador?.escopo === "instituicao"
  const voltarIndicador = `/app/indicadores/${id}`

  const {
    estado,
    carregado,
    definirFiltros,
    ordenarPor,
    definirPagina,
    definirTamanhoDePagina,
    ajustarPaginaAoTotal,
  } = useEstadoDeGrid<FiltroMetas>(
    `metas-do-indicador:${id}`,
    { busca: "", indicador_id: id, origem: "todas", situacao: "ativo" },
    "nome",
    "asc",
    20
  )
  const { data: metasData, isLoading: carregandoMetas, error: erroMetas, pesquisar } = useMetas()
  const { alterarSituacao, idEmAndamento } = useAlterarSituacaoMeta()
  const { excluir, idEmAndamento: idExcluindo } = useExcluirMeta()
  const { atualizar: atualizarMeta, isSubmitting: desvinculando } = useAtualizarMeta()
  const [idDesvinculando, setIdDesvinculando] = useState<string | null>(null)

  const [jaPesquisouMetas, setJaPesquisouMetas] = useState(false)
  const [paraInativar, setParaInativar] = useState<Meta | null>(null)
  const [paraExcluir, setParaExcluir] = useState<Meta | null>(null)

  async function executarPesquisaMetas() {
    setJaPesquisouMetas(true)
    const resultado = await pesquisar(
      { ...estado.filtros, indicador_id: id },
      { page: estado.page, page_size: estado.pageSize, sort: estado.sort, order: estado.order }
    )
    if (resultado) ajustarPaginaAoTotal(resultado.meta.total_pages)
  }

  useEffect(() => {
    if (carregado && jaPesquisouMetas) executarPesquisaMetas()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [estado.page, estado.pageSize, estado.sort, estado.order])

  useEffect(() => {
    if (jaPesquisouMetas) executarPesquisaMetas()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [estado.filtros])

  function handlePesquisarMetas(situacao: FiltroMetas["situacao"]) {
    definirFiltros({ ...estado.filtros, indicador_id: id, situacao })
    setJaPesquisouMetas(true)
  }

  function handleAlterarSituacaoMeta(item: Meta) {
    if (item.situacao === "ativo") {
      setParaInativar(item)
      return
    }
    reativarMeta(item)
  }

  async function reativarMeta(item: Meta) {
    const resultado = await alterarSituacao(item.id, "ativo", item.versao)
    if (resultado) {
      notificar.sucesso(`${item.nome} foi reativada.`)
      executarPesquisaMetas()
    }
  }

  async function confirmarInativacaoMeta() {
    if (!paraInativar) return
    const resultado = await alterarSituacao(paraInativar.id, "inativo", paraInativar.versao)
    if (resultado) {
      notificar.sucesso(`${paraInativar.nome} foi inativada.`)
      setParaInativar(null)
      executarPesquisaMetas()
    }
  }

  async function confirmarExclusaoMeta() {
    if (!paraExcluir) return
    const sucesso = await excluir(paraExcluir.id)
    if (sucesso) {
      notificar.sucesso(`${paraExcluir.nome} foi excluída.`)
      setParaExcluir(null)
      executarPesquisaMetas()
    }
  }

  async function desvincular(item: Meta) {
    setIdDesvinculando(item.id)
    const novosIndicadores = item.indicadores.filter((i) => i.id !== id).map((i) => i.id)
    const resultado = await atualizarMeta(
      item.id,
      {
        nome: item.nome,
        descricao: item.descricao,
        indicadores: novosIndicadores,
        quantidade_sugerida: item.quantidade_sugerida,
      },
      item.versao
    )
    setIdDesvinculando(null)
    if (resultado) {
      notificar.sucesso(`${item.nome} foi desvinculada deste indicador.`)
      executarPesquisaMetas()
    } else {
      notificar.erro("Não foi possível desvincular — recarregue e tente novamente.")
    }
  }

  const metas = metasData?.data ?? []

  return (
    <div className="w-full space-y-8">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Metas" },
          { rotulo: "Indicadores", href: "/app/indicadores" },
          { rotulo: indicador ? `${indicador.codigo} — ${indicador.nome}` : "Indicador" },
        ]}
      />

      <div>
        <h1 className="text-2xl font-semibold">
          {proprio ? "Editar indicador" : "Indicador do INEP"}
        </h1>

        {isLoading && (
          <div className="mt-4 max-w-3xl space-y-4">
            <div className="grid gap-4 md:grid-cols-2">
              <Skeleton className="h-16 w-full" />
              <Skeleton className="h-16 w-full" />
            </div>
            <Skeleton className="h-16 w-full" />
          </div>
        )}

        {!isLoading && error && (
          <ErrorState
            titulo="Não foi possível carregar o indicador agora."
            onTentarNovamente={recarregar}
          />
        )}

        {!isLoading && !error && indicador && proprio && (
          <div className="mt-4">
            <IndicadorForm
              key={indicador.versao}
              indicador={indicador}
              onDirtyChange={setSujo}
              onCancelar={() => solicitarSaida("/app/indicadores")}
              onConflito={recarregar}
              onSalvo={() => recarregar()}
            />
          </div>
        )}

        {!isLoading && !error && indicador && !proprio && (
          <div className="mt-4 max-w-3xl space-y-2 rounded-md border p-4">
            <div className="flex items-center gap-2">
              <p className="font-medium">
                {indicador.codigo} — {indicador.nome}
              </p>
              <StatusBadge label="Do INEP" variant="secondary" />
            </div>
            {indicador.referencia_instrumento && (
              <p className="text-sm text-muted-foreground">
                Referência: {indicador.referencia_instrumento}
              </p>
            )}
            {indicador.descricao && <p className="text-sm">{indicador.descricao}</p>}
            <p className="text-sm text-muted-foreground">
              🔒 Este indicador é mantido pelo Administrador do Sistema — somente leitura aqui. As
              metas abaixo são da sua instituição e você pode geri-las normalmente.
            </p>
          </div>
        )}
      </div>

      {!isLoading && !error && indicador && (
        <div className="space-y-4 border-t pt-6">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold">Metas atendidas por este indicador</h2>
            <LoadingButton
              render={
                <Link
                  href={`/app/metas/novo?indicador_id=${indicador.id}&voltar=${encodeURIComponent(voltarIndicador)}`}
                />
              }
            >
              <PlusIcon aria-hidden="true" /> Criar meta com este indicador
              <IndicadorDeNavegacao />
            </LoadingButton>
          </div>

          <div className="flex flex-wrap items-end gap-3">
            <Field className="w-48">
              <FieldLabel htmlFor="metas-indicador-situacao">Situação</FieldLabel>
              <FieldContent>
                <SelectComRotulo
                  id="metas-indicador-situacao"
                  value={estado.filtros.situacao}
                  onValueChange={(v) => handlePesquisarMetas(v as FiltroMetas["situacao"])}
                  itens={[
                    { value: "ativo", label: "Ativas" },
                    { value: "inativo", label: "Inativas" },
                    { value: "todas", label: "Todas" },
                  ]}
                />
              </FieldContent>
            </Field>
            <LoadingButton
              loading={carregandoMetas}
              loadingText="Pesquisando..."
              onClick={() => handlePesquisarMetas(estado.filtros.situacao)}
            >
              Pesquisar
            </LoadingButton>
          </div>

          {!jaPesquisouMetas && (
            <EmptyState
              icon={ClipboardListIcon}
              titulo="Clique em Pesquisar para ver as metas que atendem este indicador."
            />
          )}

          {jaPesquisouMetas && carregandoMetas && (
            <SkeletonTable colunas={["Nome", "Indicadores", "Situação", "Planos", "Ações"]} />
          )}

          {jaPesquisouMetas && !carregandoMetas && erroMetas && (
            <ErrorState
              titulo="Não foi possível carregar as metas agora."
              onTentarNovamente={executarPesquisaMetas}
            />
          )}

          {jaPesquisouMetas && !carregandoMetas && !erroMetas && metas.length === 0 && (
            <EmptyState
              icon={ClipboardListIcon}
              titulo="Nenhuma meta atende este indicador ainda."
              descricao='Use "Criar meta com este indicador" acima para cadastrar a primeira.'
            />
          )}

          {jaPesquisouMetas && !carregandoMetas && !erroMetas && metas.length > 0 && (
            <div className="space-y-4" aria-busy={carregandoMetas}>
              <p className="sr-only" aria-live="polite">
                {metasData?.meta.total} metas encontradas.
              </p>
              <div className="hidden md:block">
                <MetaTable
                  itens={metas}
                  sort={estado.sort}
                  order={estado.order}
                  onOrdenar={ordenarPor}
                  onAlterarSituacao={handleAlterarSituacaoMeta}
                  onExcluir={setParaExcluir}
                  idAlterandoSituacao={idEmAndamento ?? idExcluindo}
                  voltar={voltarIndicador}
                  indicadorContextoId={indicador.id}
                  onDesvincular={desvincular}
                  idDesvinculando={desvinculando ? idDesvinculando : null}
                />
              </div>
              <div className="grid gap-3 md:hidden">
                {metas.map((item) => (
                  <MetaCardMobile
                    key={item.id}
                    item={item}
                    onAlterarSituacao={handleAlterarSituacaoMeta}
                    onExcluir={setParaExcluir}
                    idAlterandoSituacao={idEmAndamento ?? idExcluindo}
                    voltar={voltarIndicador}
                    indicadorContextoId={indicador.id}
                    onDesvincular={desvincular}
                    idDesvinculando={desvinculando ? idDesvinculando : null}
                  />
                ))}
              </div>
              {metasData && (
                <Paginacao
                  page={metasData.meta.page}
                  pageSize={metasData.meta.page_size}
                  total={metasData.meta.total}
                  totalPages={metasData.meta.total_pages}
                  onPageChange={definirPagina}
                  onPageSizeChange={definirTamanhoDePagina}
                />
              )}
              <p className="text-sm text-muted-foreground">
                Desvincular remove este indicador daquela meta (a meta continua existindo, com os
                demais indicadores). Excluir apaga a meta inteira — só é possível sem plano
                vinculado.
              </p>
            </div>
          )}
        </div>
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

      <InativarMetaDialog
        meta={paraInativar}
        onOpenChange={(aberto) => !aberto && setParaInativar(null)}
        confirmando={idEmAndamento === paraInativar?.id}
        onConfirmar={confirmarInativacaoMeta}
      />

      <ExcluirMetaDialog
        meta={paraExcluir}
        onOpenChange={(aberto) => !aberto && setParaExcluir(null)}
        confirmando={idExcluindo === paraExcluir?.id}
        onConfirmar={confirmarExclusaoMeta}
      />
    </div>
  )
}
