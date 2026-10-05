"use client"

import { useEffect, useId, useRef, useState } from "react"
import { useDirtyState } from "@/components/shared/hooks/use-dirty-state"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Field, FieldContent, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { focarPrimeiroCampoComErro } from "@/lib/foco"
import { obrigatorio } from "@/lib/validacao"
import { useAtualizarIndicador } from "../hooks/use-atualizar-indicador"
import { useCriarIndicador } from "../hooks/use-criar-indicador"
import type { Indicador } from "../types"

interface IndicadorFormProps {
  indicador: Indicador | null
  onSalvo: (indicador: Indicador, criado: boolean) => void
  onCancelar: () => void
  onDirtyChange?: (sujo: boolean) => void
  onConflito?: () => void
  criarOutro?: boolean
}

export function IndicadorForm({
  indicador,
  onSalvo,
  onCancelar,
  onDirtyChange,
  onConflito,
  criarOutro = false,
}: IndicadorFormProps) {
  const modoEdicao = !!indicador
  const idBase = useId()
  const primeiroCampoRef = useRef<HTMLInputElement>(null)
  const formRef = useRef<HTMLFormElement>(null)

  const { valores, definirCampo, sujo, reiniciar } = useDirtyState({
    codigo: indicador?.codigo ?? "",
    nome: indicador?.nome ?? "",
    descricao: indicador?.descricao ?? "",
  })

  const [erroCodigo, setErroCodigo] = useState<string | undefined>()
  const [erroNome, setErroNome] = useState<string | undefined>()
  const [conflito, setConflito] = useState(false)

  const {
    criar,
    isSubmitting: criando,
    error: erroCriar,
    limparErro: limparErroCriar,
  } = useCriarIndicador()
  const {
    atualizar,
    isSubmitting: atualizando,
    error: erroAtualizar,
    limparErro: limparErroAtualizar,
  } = useAtualizarIndicador()

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
    if (erro?.code === "CONFLITO_DE_VERSAO") setConflito(true)
    if (erro?.code === "CODIGO_INDICADOR_DUPLICADO") {
      setErroCodigo("Já existe um indicador com este código nesta instituição.")
    }
  }, [erro])

  function limparFormulario() {
    reiniciar()
    setErroCodigo(undefined)
    setErroNome(undefined)
    setConflito(false)
    limparErroCriar()
    limparErroAtualizar()
    requestAnimationFrame(() => primeiroCampoRef.current?.focus())
  }

  async function submeter(continuarCriando: boolean) {
    const semCodigo = obrigatorio(valores.codigo, "O código é obrigatório.")
    const semNome = obrigatorio(valores.nome, "O nome é obrigatório.")
    setErroCodigo(semCodigo)
    setErroNome(semNome)
    if (semCodigo || semNome) {
      focarPrimeiroCampoComErro(formRef.current)
      return
    }

    const input = { codigo: valores.codigo, nome: valores.nome, descricao: valores.descricao }

    if (modoEdicao && indicador) {
      const resultado = await atualizar(indicador.id, input, indicador.versao)
      if (resultado) {
        notificar.sucesso("Indicador atualizado com sucesso.")
        onSalvo(resultado, false)
      }
      return
    }

    const resultado = await criar(input)
    if (resultado) {
      notificar.sucesso("Indicador cadastrado com sucesso.")
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

        <Field data-invalid={!!erroCodigo}>
          <FieldLabel htmlFor={`${idBase}-codigo`}>Código</FieldLabel>
          <FieldContent>
            <Input
              id={`${idBase}-codigo`}
              ref={primeiroCampoRef}
              value={valores.codigo}
              onChange={(e) => definirCampo("codigo", e.target.value)}
              disabled={salvando}
              aria-invalid={!!erroCodigo}
              aria-describedby={erroCodigo ? `${idBase}-codigo-erro` : undefined}
            />
            {erroCodigo && <FieldError id={`${idBase}-codigo-erro`}>{erroCodigo}</FieldError>}
          </FieldContent>
        </Field>

        <Field data-invalid={!!erroNome}>
          <FieldLabel htmlFor={`${idBase}-nome`}>Nome</FieldLabel>
          <FieldContent>
            <Input
              id={`${idBase}-nome`}
              value={valores.nome}
              onChange={(e) => definirCampo("nome", e.target.value)}
              disabled={salvando}
              aria-invalid={!!erroNome}
              aria-describedby={erroNome ? `${idBase}-nome-erro` : undefined}
            />
            {erroNome && <FieldError id={`${idBase}-nome-erro`}>{erroNome}</FieldError>}
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
