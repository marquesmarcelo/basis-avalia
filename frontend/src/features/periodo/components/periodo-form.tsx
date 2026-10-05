"use client"

import { useEffect, useId, useRef, useState } from "react"
import { useDirtyState } from "@/components/shared/hooks/use-dirty-state"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Field, FieldContent, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { focarPrimeiroCampoComErro } from "@/lib/foco"
import { formatarDataPura, hojeDataPura } from "@/lib/formato"
import { obrigatorio } from "@/lib/validacao"
import { useAtualizarPeriodo } from "../hooks/use-atualizar-periodo"
import { useCriarPeriodo } from "../hooks/use-criar-periodo"
import { calcularSituacao, ROTULO_SITUACAO } from "../lib/situacao"
import type { Periodo } from "../types"

interface PeriodoFormProps {
  periodo: Periodo | null
  onSalvo: (periodo: Periodo) => void
  onCancelar: () => void
  onDirtyChange?: (sujo: boolean) => void
  onConflito?: () => void
  criarOutro?: boolean
}

export function PeriodoForm({
  periodo,
  onSalvo,
  onCancelar,
  onDirtyChange,
  onConflito,
  criarOutro = false,
}: PeriodoFormProps) {
  const modoEdicao = !!periodo
  const idBase = useId()
  const primeiroCampoRef = useRef<HTMLInputElement>(null)
  const formRef = useRef<HTMLFormElement>(null)
  const versaoConhecidaRef = useRef(periodo?.versao)

  const { valores, definirCampo, sujo, reiniciar } = useDirtyState({
    nome: periodo?.nome ?? "",
    data_inicio: periodo?.data_inicio ?? "",
    data_fim: periodo?.data_fim ?? "",
  })

  const [erroNome, setErroNome] = useState<string | undefined>()
  const [erroDatas, setErroDatas] = useState<string | undefined>()
  const [conflito, setConflito] = useState(false)

  const {
    criar,
    isSubmitting: criando,
    error: erroCriar,
    limparErro: limparErroCriar,
  } = useCriarPeriodo()
  const {
    atualizar,
    isSubmitting: atualizando,
    error: erroAtualizar,
    limparErro: limparErroAtualizar,
  } = useAtualizarPeriodo()

  const salvando = criando || atualizando
  const erro = erroCriar ?? erroAtualizar

  useEffect(() => {
    onDirtyChange?.(sujo)
  }, [sujo, onDirtyChange])

  useEffect(() => {
    requestAnimationFrame(() => primeiroCampoRef.current?.focus())
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // Depois de "Recarregar dados" o pai reexecuta o GET e a prop `periodo`
  // chega com a versão nova — sem isto, o formulário continuaria mostrando
  // os valores antigos por cima do alerta de conflito já fechado.
  useEffect(() => {
    if (periodo && periodo.versao !== versaoConhecidaRef.current) {
      versaoConhecidaRef.current = periodo.versao
      reiniciar()
      setConflito(false)
    }
  }, [periodo, reiniciar])

  useEffect(() => {
    if (erro?.code === "CONFLITO_DE_VERSAO") setConflito(true)
    if (erro?.code === "NOME_PERIODO_DUPLICADO") {
      setErroNome("Já existe um período com este nome nesta instituição.")
    }
    if (erro?.code === "PERIODO_DATAS_INVALIDAS") {
      setErroDatas("A data de fim precisa ser igual ou posterior à data de início.")
    }
  }, [erro])

  const reabreOPeriodo =
    !!periodo && periodo.situacao === "encerrado" && valores.data_fim > periodo.data_fim

  const hojeISO = hojeDataPura()
  const situacaoCalculada =
    valores.data_inicio && valores.data_fim
      ? calcularSituacao(valores.data_inicio, valores.data_fim, hojeISO)
      : null

  function limparFormulario() {
    reiniciar()
    setErroNome(undefined)
    setErroDatas(undefined)
    setConflito(false)
    limparErroCriar()
    limparErroAtualizar()
    requestAnimationFrame(() => primeiroCampoRef.current?.focus())
  }

  async function submeter(continuarCriando: boolean) {
    const semNome = obrigatorio(valores.nome, "Informe um nome para o período.")
    setErroNome(semNome)
    if (valores.data_fim && valores.data_inicio && valores.data_fim < valores.data_inicio) {
      setErroDatas("A data de fim precisa ser igual ou posterior à data de início.")
      focarPrimeiroCampoComErro(formRef.current)
      return
    }
    setErroDatas(undefined)
    if (semNome) {
      focarPrimeiroCampoComErro(formRef.current)
      return
    }

    const input = {
      nome: valores.nome,
      data_inicio: valores.data_inicio,
      data_fim: valores.data_fim,
    }

    if (modoEdicao && periodo) {
      const resultado = await atualizar(periodo.id, input, periodo.versao)
      if (resultado) {
        notificar.sucesso("Período atualizado.")
        onSalvo(resultado)
      }
      return
    }

    const resultado = await criar(input)
    if (resultado) {
      notificar.sucesso("Período cadastrado.")
      if (continuarCriando) {
        limparFormulario()
      } else {
        onSalvo(resultado)
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
                onClick={() => onConflito?.()}
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

        <Field>
          <FieldLabel htmlFor={`${idBase}-inicio`}>Data de início</FieldLabel>
          <FieldContent>
            <Input
              id={`${idBase}-inicio`}
              type="date"
              value={valores.data_inicio}
              onChange={(e) => definirCampo("data_inicio", e.target.value)}
              disabled={salvando}
            />
          </FieldContent>
        </Field>

        <Field data-invalid={!!erroDatas}>
          <FieldLabel htmlFor={`${idBase}-fim`}>Data de fim</FieldLabel>
          <FieldContent>
            <Input
              id={`${idBase}-fim`}
              type="date"
              value={valores.data_fim}
              onChange={(e) => definirCampo("data_fim", e.target.value)}
              disabled={salvando}
              aria-invalid={!!erroDatas}
              aria-describedby={`${idBase}-fim-aviso${erroDatas ? ` ${idBase}-fim-erro` : ""}`}
            />
            <p id={`${idBase}-fim-aviso`} className="text-sm text-muted-foreground">
              O dia inteiro da data de fim conta — entregas são aceitas até 23h59 desse dia, no
              horário de Brasília.
            </p>
            {erroDatas && <FieldError id={`${idBase}-fim-erro`}>{erroDatas}</FieldError>}
          </FieldContent>
        </Field>

        {situacaoCalculada && (
          <Field className="md:col-span-full">
            <FieldLabel>Situação</FieldLabel>
            <FieldContent>
              <div className="flex items-center gap-2">
                <StatusBadge
                  label={ROTULO_SITUACAO[situacaoCalculada]}
                  variant={situacaoCalculada === "aberto" ? "default" : "secondary"}
                />
                <span className="text-sm text-muted-foreground">
                  Não é um campo — calculada a partir das datas acima, comparadas com hoje (
                  {formatarDataPura(hojeISO)}).
                </span>
              </div>
            </FieldContent>
          </Field>
        )}

        {reabreOPeriodo && (
          <Alert className="md:col-span-full">
            <AlertDescription>
              Prorrogar reabre este período e os planos vigentes dele, que voltam a aceitar entrega.
            </AlertDescription>
          </Alert>
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
