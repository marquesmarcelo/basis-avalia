"use client"

import { useEffect, useId, useRef, useState } from "react"
import { useDirtyState } from "@/components/shared/hooks/use-dirty-state"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Field, FieldContent, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { focarPrimeiroCampoComErro } from "@/lib/foco"
import { formatarData } from "@/lib/formato"
import { obrigatorio } from "@/lib/validacao"
import { useAtualizarInstituicao } from "../hooks/use-atualizar-instituicao"
import { useCriarInstituicao } from "../hooks/use-criar-instituicao"
import type { Instituicao } from "../types"

interface InstituicaoFormProps {
  instituicao: Instituicao | null
  onSalvo: (instituicao: Instituicao, criada: boolean) => void
  onCancelar: () => void
  onDirtyChange?: (sujo: boolean) => void
  onConflito?: () => void
  criarOutro?: boolean
}

export function InstituicaoForm({
  instituicao,
  onSalvo,
  onCancelar,
  onDirtyChange,
  onConflito,
  criarOutro = false,
}: InstituicaoFormProps) {
  const modoEdicao = !!instituicao
  const idNome = useId()
  const primeiroCampoRef = useRef<HTMLInputElement>(null)
  const formRef = useRef<HTMLFormElement>(null)

  const { valores, definirCampo, sujo, reiniciar } = useDirtyState({
    nome: instituicao?.nome ?? "",
    sigla: instituicao?.sigla ?? "",
    codigo_emec: instituicao?.codigo_emec ?? "",
  })

  const [erroNome, setErroNome] = useState<string | undefined>()
  const [erroSigla, setErroSigla] = useState<string | undefined>()
  const [conflito, setConflito] = useState(false)

  const {
    criar,
    isSubmitting: criando,
    error: erroCriar,
    limparErro: limparErroCriar,
  } = useCriarInstituicao()
  const {
    atualizar,
    isSubmitting: atualizando,
    error: erroAtualizar,
    limparErro: limparErroAtualizar,
  } = useAtualizarInstituicao()

  const salvando = criando || atualizando
  const erro = erroCriar ?? erroAtualizar

  useEffect(() => {
    onDirtyChange?.(sujo)
  }, [sujo, onDirtyChange])

  useEffect(() => {
    requestAnimationFrame(() => primeiroCampoRef.current?.focus())
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (erro?.code === "CONFLITO_DE_VERSAO") {
      setConflito(true)
    }
  }, [erro])

  function limparFormulario() {
    reiniciar()
    setErroNome(undefined)
    setErroSigla(undefined)
    setConflito(false)
    limparErroCriar()
    limparErroAtualizar()
    requestAnimationFrame(() => primeiroCampoRef.current?.focus())
  }

  async function submeter(continuarCriando: boolean) {
    const semNome = obrigatorio(valores.nome, "O nome é obrigatório.")
    const semSigla = obrigatorio(valores.sigla, "A sigla é obrigatória.")
    setErroNome(semNome)
    setErroSigla(semSigla)
    if (semNome || semSigla) {
      focarPrimeiroCampoComErro(formRef.current)
      return
    }

    const input = { nome: valores.nome, sigla: valores.sigla, codigo_emec: valores.codigo_emec }

    if (modoEdicao && instituicao) {
      const resultado = await atualizar(instituicao.id, input, instituicao.versao)
      if (resultado) {
        notificar.sucesso("Instituição atualizada com sucesso.")
        onSalvo(resultado, false)
      }
      return
    }

    const resultado = await criar(input)
    if (resultado) {
      notificar.sucesso("Instituição cadastrada com sucesso.")
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
          <FieldLabel htmlFor={`${idNome}-nome`}>Nome</FieldLabel>
          <FieldContent>
            <Input
              id={`${idNome}-nome`}
              ref={primeiroCampoRef}
              value={valores.nome}
              onChange={(e) => definirCampo("nome", e.target.value)}
              disabled={salvando}
              aria-invalid={!!erroNome}
              aria-describedby={erroNome ? `${idNome}-nome-erro` : undefined}
            />
            {erroNome && <FieldError id={`${idNome}-nome-erro`}>{erroNome}</FieldError>}
          </FieldContent>
        </Field>

        <Field data-invalid={!!erroSigla}>
          <FieldLabel htmlFor={`${idNome}-sigla`}>Sigla</FieldLabel>
          <FieldContent>
            <Input
              id={`${idNome}-sigla`}
              value={valores.sigla}
              onChange={(e) => definirCampo("sigla", e.target.value)}
              disabled={salvando}
              aria-invalid={!!erroSigla}
              aria-describedby={erroSigla ? `${idNome}-sigla-erro` : undefined}
            />
            {erroSigla && <FieldError id={`${idNome}-sigla-erro`}>{erroSigla}</FieldError>}
          </FieldContent>
        </Field>

        <Field>
          <FieldLabel htmlFor={`${idNome}-emec`}>Código e-MEC</FieldLabel>
          <FieldContent>
            <Input
              id={`${idNome}-emec`}
              value={valores.codigo_emec}
              onChange={(e) => definirCampo("codigo_emec", e.target.value)}
              disabled={salvando}
            />
          </FieldContent>
        </Field>

        {modoEdicao && instituicao && (
          <p className="text-sm text-muted-foreground md:col-span-full">
            Situação atual: {instituicao.situacao === "ativa" ? "Ativa" : "Inativa"} · Cadastrado em{" "}
            {formatarData(instituicao.criado_em)}
          </p>
        )}
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
