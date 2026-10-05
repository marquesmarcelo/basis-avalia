"use client"

import { useEffect, useId, useMemo, useRef, useState } from "react"
import {
  ComboboxEntidadeMultipla,
  type ItemComboboxMultiplo,
} from "@/components/shared/forms/combobox-entidade-multipla"
import { useDirtyState } from "@/components/shared/hooks/use-dirty-state"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Field, FieldContent, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { useIndicador } from "@/features/indicador/hooks/use-indicador"
import { useIndicadoresSugestoes } from "@/features/indicador/hooks/use-indicadores-sugestoes"
import { focarPrimeiroCampoComErro } from "@/lib/foco"
import { obrigatorio } from "@/lib/validacao"
import { useAtualizarMeta } from "../hooks/use-atualizar-meta"
import { useCriarMeta } from "../hooks/use-criar-meta"
import type { Meta } from "../types"

const MAX_INDICADORES = 5

interface MetaFormProps {
  meta: Meta | null
  indicadorPreSelecionadoId?: string
  onSalvo: (meta: Meta, criada: boolean) => void
  onCancelar: () => void
  onDirtyChange?: (sujo: boolean) => void
  onConflito?: () => void
  criarOutro?: boolean
}

export function MetaForm({
  meta,
  indicadorPreSelecionadoId,
  onSalvo,
  onCancelar,
  onDirtyChange,
  onConflito,
  criarOutro = false,
}: MetaFormProps) {
  const modoEdicao = !!meta
  const idBase = useId()
  const primeiroCampoRef = useRef<HTMLInputElement>(null)
  const formRef = useRef<HTMLFormElement>(null)

  const {
    valores,
    definirCampo,
    sujo: sujoTexto,
    reiniciar,
  } = useDirtyState({
    nome: meta?.nome ?? "",
    descricao: meta?.descricao ?? "",
    quantidade_sugerida: meta?.quantidade_sugerida != null ? String(meta.quantidade_sugerida) : "",
  })
  const valorInicialIndicadores = useMemo(
    () =>
      meta?.indicadores.map((i) => i.id) ??
      (indicadorPreSelecionadoId ? [indicadorPreSelecionadoId] : []),
    [meta, indicadorPreSelecionadoId]
  )
  const [indicadoresSelecionados, setIndicadoresSelecionados] =
    useState<string[]>(valorInicialIndicadores)

  const [erroNome, setErroNome] = useState<string | undefined>()
  const [erroIndicadores, setErroIndicadores] = useState<string | undefined>()
  const [erroQuantidade, setErroQuantidade] = useState<string | undefined>()
  const [conflito, setConflito] = useState(false)

  const {
    itens,
    isLoading: carregandoIndicadores,
    error: erroIndicadoresCarga,
    buscar,
  } = useIndicadoresSugestoes()
  // O indicador pré-selecionado pode não estar na primeira página de
  // sugestões — busca ele à parte para o combobox mostrar o rótulo, não
  // só o id (§7.3 da conversão de página: o campo continua editável, mas
  // precisa nascer visível).
  const { data: indicadorPreSelecionado } = useIndicador(indicadorPreSelecionadoId ?? "")

  const {
    criar,
    isSubmitting: criando,
    error: erroCriar,
    limparErro: limparErroCriar,
  } = useCriarMeta()
  const {
    atualizar,
    isSubmitting: atualizando,
    error: erroAtualizar,
    limparErro: limparErroAtualizar,
  } = useAtualizarMeta()

  const salvando = criando || atualizando
  const erro = erroCriar ?? erroAtualizar

  const sujo =
    sujoTexto || JSON.stringify(indicadoresSelecionados) !== JSON.stringify(valorInicialIndicadores)

  useEffect(() => {
    onDirtyChange?.(sujo)
  }, [sujo, onDirtyChange])

  useEffect(() => {
    buscar("")
    requestAnimationFrame(() => primeiroCampoRef.current?.focus())
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (erro?.code === "CONFLITO_DE_VERSAO") setConflito(true)
    if (erro?.code === "NOME_META_DUPLICADO") {
      setErroNome("Já existe uma meta com este nome nesta instituição.")
    }
    if (erro?.code === "INDICADOR_OBRIGATORIO") {
      setErroIndicadores("Selecione pelo menos um indicador.")
    }
    if (erro?.code === "INDICADORES_ACIMA_DO_LIMITE") {
      setErroIndicadores("Uma meta pode ter no máximo 5 indicadores.")
    }
    if (erro?.code === "QUANTIDADE_INVALIDA") {
      setErroQuantidade("Informe uma quantidade de 1 ou mais, ou deixe em branco.")
    }
  }, [erro])

  const itensCombobox = useMemo<ItemComboboxMultiplo[]>(() => {
    const base: ItemComboboxMultiplo[] = itens.map((i) => ({
      value: i.id,
      label: `${i.codigo} — ${i.nome}`,
      badge: i.escopo === "plataforma" ? "Do INEP" : "Próprio",
      badgeVariant: i.escopo === "plataforma" ? "secondary" : "outline",
      secundario: i.escopo === "plataforma" ? i.referencia_instrumento : undefined,
    }))
    if (indicadorPreSelecionado && !base.some((b) => b.value === indicadorPreSelecionado.id)) {
      base.unshift({
        value: indicadorPreSelecionado.id,
        label: `${indicadorPreSelecionado.codigo} — ${indicadorPreSelecionado.nome}`,
        badge: indicadorPreSelecionado.escopo === "plataforma" ? "Do INEP" : "Próprio",
        badgeVariant: indicadorPreSelecionado.escopo === "plataforma" ? "secondary" : "outline",
        secundario:
          indicadorPreSelecionado.escopo === "plataforma"
            ? (indicadorPreSelecionado.referencia_instrumento ?? undefined)
            : undefined,
      })
    }
    return base
  }, [itens, indicadorPreSelecionado])

  function limparFormulario() {
    reiniciar()
    setIndicadoresSelecionados(indicadorPreSelecionadoId ? [indicadorPreSelecionadoId] : [])
    setErroNome(undefined)
    setErroIndicadores(undefined)
    setErroQuantidade(undefined)
    setConflito(false)
    limparErroCriar()
    limparErroAtualizar()
    requestAnimationFrame(() => primeiroCampoRef.current?.focus())
  }

  async function submeter(continuarCriando: boolean) {
    const semNome = obrigatorio(valores.nome, "O nome é obrigatório.")
    setErroNome(semNome)
    let semIndicadores: string | undefined
    if (indicadoresSelecionados.length === 0) {
      semIndicadores = "Selecione pelo menos um indicador."
      setErroIndicadores(semIndicadores)
    } else {
      setErroIndicadores(undefined)
    }
    let quantidadeSugerida: number | null = null
    let erroQtd: string | undefined
    if (valores.quantidade_sugerida.trim() !== "") {
      const numero = Number(valores.quantidade_sugerida)
      if (!Number.isInteger(numero) || numero < 1) {
        erroQtd = "Informe uma quantidade de 1 ou mais, ou deixe em branco."
      } else {
        quantidadeSugerida = numero
      }
    }
    setErroQuantidade(erroQtd)
    if (semNome || semIndicadores || erroQtd) {
      focarPrimeiroCampoComErro(formRef.current)
      return
    }

    const input = {
      nome: valores.nome,
      descricao: valores.descricao,
      indicadores: indicadoresSelecionados,
      quantidade_sugerida: quantidadeSugerida,
    }

    if (modoEdicao && meta) {
      const resultado = await atualizar(meta.id, input, meta.versao)
      if (resultado) {
        notificar.sucesso("Meta atualizada com sucesso.")
        onSalvo(resultado, false)
      }
      return
    }

    const resultado = await criar(input)
    if (resultado) {
      notificar.sucesso("Meta cadastrada com sucesso.")
      if (continuarCriando) {
        limparFormulario()
      } else {
        onSalvo(resultado, true)
      }
    }
  }

  return (
    <form
      ref={formRef}
      onSubmit={(e) => {
        e.preventDefault()
        submeter(false)
      }}
      aria-busy={salvando}
      className="max-w-3xl"
      onKeyDown={(e) => {
        if ((e.ctrlKey || e.metaKey) && e.key === "s") {
          e.preventDefault()
          submeter(false)
        }
      }}
    >
      <FieldGroup className="md:grid md:grid-cols-2 md:gap-4">
        {conflito && (
          <Alert variant="destructive" role="alert" className="md:col-span-full">
            <AlertTitle>Registro alterado</AlertTitle>
            <AlertDescription>
              Este registro foi alterado por outro usuário enquanto você editava.
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="mt-2"
                onClick={onConflito}
              >
                Recarregar dados
              </Button>
            </AlertDescription>
          </Alert>
        )}

        <Field data-invalid={!!erroNome}>
          <FieldLabel htmlFor={`${idBase}-nome`}>Nome</FieldLabel>
          <FieldContent>
            <Input
              id={`${idBase}-nome`}
              ref={primeiroCampoRef}
              value={valores.nome}
              onChange={(e) => definirCampo("nome", e.target.value)}
              disabled={salvando}
              aria-invalid={!!erroNome}
              aria-describedby={erroNome ? `${idBase}-nome-erro` : undefined}
            />
            {erroNome && <FieldError id={`${idBase}-nome-erro`}>{erroNome}</FieldError>}
          </FieldContent>
        </Field>

        <Field data-invalid={!!erroQuantidade}>
          <FieldLabel htmlFor={`${idBase}-quantidade`}>Quantidade sugerida (opcional)</FieldLabel>
          <FieldContent>
            <Input
              id={`${idBase}-quantidade`}
              type="number"
              min={1}
              step={1}
              value={valores.quantidade_sugerida}
              onChange={(e) => definirCampo("quantidade_sugerida", e.target.value)}
              disabled={salvando}
              aria-invalid={!!erroQuantidade}
              aria-describedby={`${idBase}-quantidade-nota`}
            />
            <p id={`${idBase}-quantidade-nota`} className="text-sm text-muted-foreground">
              O plano herda este valor ao adicionar a meta — o PI pode ajustar por curso depois.
              Quem decide a apuração é sempre a quantidade do item do plano, nunca esta sugestão.
            </p>
            {erroQuantidade && <FieldError>{erroQuantidade}</FieldError>}
          </FieldContent>
        </Field>

        <Field className="md:col-span-full">
          <FieldLabel htmlFor={`${idBase}-descricao`}>Descrição</FieldLabel>
          <FieldContent>
            <Textarea
              id={`${idBase}-descricao`}
              value={valores.descricao}
              onChange={(e) => definirCampo("descricao", e.target.value)}
              disabled={salvando}
            />
          </FieldContent>
        </Field>

        <Field className="md:col-span-full" data-invalid={!!erroIndicadores}>
          <FieldLabel htmlFor={`${idBase}-indicadores`}>Indicadores (1 a 5)</FieldLabel>
          <FieldContent>
            <ComboboxEntidadeMultipla
              id={`${idBase}-indicadores`}
              itens={itensCombobox}
              value={indicadoresSelecionados}
              onValueChange={setIndicadoresSelecionados}
              max={MAX_INDICADORES}
              carregando={carregandoIndicadores}
              erro={!!erroIndicadoresCarga}
              onTentarNovamente={() => buscar("")}
              placeholder="Buscar indicador..."
              disabled={salvando}
            />
            {erroIndicadores && (
              <FieldError id={`${idBase}-indicadores-erro`}>{erroIndicadores}</FieldError>
            )}
          </FieldContent>
        </Field>

        <p className="text-sm text-muted-foreground md:col-span-full">
          Uma entrega desta meta atende a todos os indicadores selecionados. A quantidade exigida
          continua sendo a do item do plano, não muda com o número de indicadores.
        </p>
      </FieldGroup>

      <div className="mt-6 flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onCancelar} disabled={salvando}>
          Cancelar
        </Button>
        {criarOutro && !modoEdicao && (
          <LoadingButton
            type="button"
            variant="outline"
            loading={salvando}
            loadingText="Salvando..."
            onClick={() => submeter(true)}
          >
            Salvar e cadastrar outro
          </LoadingButton>
        )}
        <LoadingButton type="submit" loading={salvando} loadingText="Salvando...">
          Salvar
        </LoadingButton>
      </div>
    </form>
  )
}
