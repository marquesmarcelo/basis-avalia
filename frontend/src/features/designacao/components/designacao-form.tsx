"use client"

import { InfoIcon } from "lucide-react"
import { useEffect, useId, useMemo, useRef, useState } from "react"
import { ComboboxEntidade } from "@/components/shared/forms/combobox-entidade"
import { useDirtyState } from "@/components/shared/hooks/use-dirty-state"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { useEu } from "@/features/auth/hooks/use-eu"
import { focarPrimeiroCampoComErro } from "@/lib/foco"
import { formatarDataPura } from "@/lib/formato"
import { obrigatorio } from "@/lib/validacao"
import { useAtualizarDesignacao } from "../hooks/use-atualizar-designacao"
import { useCandidatosDesignacao } from "../hooks/use-candidatos-designacao"
import { useCriarDesignacao } from "../hooks/use-criar-designacao"
import type { Designacao } from "../types"

interface DesignacaoFormProps {
  cursoId: string
  designacao: Designacao | null
  onSalvo: (item: Designacao) => void
  onCancelar: () => void
  onDirtyChange?: (sujo: boolean) => void
  onConflito?: () => void
  criarOutro?: boolean
}

export function DesignacaoForm({
  cursoId,
  designacao,
  onSalvo,
  onCancelar,
  onDirtyChange,
  onConflito,
  criarOutro = false,
}: DesignacaoFormProps) {
  const modoEdicao = !!designacao
  const editavel = !modoEdicao || designacao?.situacao === "futura"
  const idBase = useId()
  const formRef = useRef<HTMLFormElement>(null)
  const { data: eu } = useEu()

  const {
    itens,
    isLoading: carregandoCandidatos,
    error: erroCandidatos,
    buscar,
  } = useCandidatosDesignacao()
  const {
    criar,
    isSubmitting: criando,
    error: erroCriar,
    limparErro: limparErroCriar,
  } = useCriarDesignacao(cursoId)
  const {
    atualizar,
    isSubmitting: atualizando,
    error: erroAtualizar,
    limparErro: limparErroAtualizar,
  } = useAtualizarDesignacao()

  const salvando = criando || atualizando
  const erro = erroCriar ?? erroAtualizar

  const { valores, definirCampo, sujo, reiniciar } = useDirtyState({
    coordenador_id: designacao?.coordenador.id ?? "",
    portaria: designacao?.portaria ?? "",
    data_inicio: designacao?.data_inicio ?? "",
    data_fim: designacao?.data_fim ?? "",
  })
  const [erroCampo, setErroCampo] = useState<Record<string, string | undefined>>({})
  const [conflito, setConflito] = useState(false)

  useEffect(() => {
    onDirtyChange?.(sujo)
  }, [sujo, onDirtyChange])

  useEffect(() => {
    if (editavel) buscar("")
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (erro?.code === "CONFLITO_DE_VERSAO") setConflito(true)
  }, [erro])

  function limparFormulario() {
    reiniciar()
    setErroCampo({})
    setConflito(false)
    limparErroCriar()
    limparErroAtualizar()
  }

  const candidatoSelecionado = useMemo(
    () => itens.find((c) => c.id === valores.coordenador_id) ?? null,
    [itens, valores.coordenador_id]
  )
  const itensCombobox = useMemo(
    () =>
      itens.map((c) => {
        const jaCoordena = c.perfis.includes("coordenador_curso")
        const ehPI = c.perfis.includes("pesquisador_institucional")
        const badge = jaCoordena ? "Já coordena" : ehPI ? "PI" : "Professor"
        return { value: c.id, label: c.nome, badge }
      }),
    [itens]
  )
  const ehAutodesignacao = !modoEdicao && !!eu && candidatoSelecionado?.id === eu.id
  const tambemEhPI =
    !modoEdicao && !!candidatoSelecionado?.perfis.includes("pesquisador_institucional")

  async function submeter(continuarCriando: boolean) {
    const proximoErro: Record<string, string | undefined> = {
      portaria: obrigatorio(valores.portaria, "A portaria é obrigatória."),
    }
    if (editavel) {
      proximoErro.coordenador_id = obrigatorio(valores.coordenador_id, "Selecione o coordenador.")
      proximoErro.data_inicio = obrigatorio(valores.data_inicio, "A data de início é obrigatória.")
    }
    if (valores.data_fim && valores.data_inicio && valores.data_fim < valores.data_inicio) {
      proximoErro.data_fim = "A data de fim não pode ser anterior à data de início."
    }
    setErroCampo(proximoErro)
    if (Object.values(proximoErro).some(Boolean)) {
      focarPrimeiroCampoComErro(formRef.current)
      return
    }

    if (modoEdicao && designacao) {
      const resultado = await atualizar(
        designacao.id,
        {
          coordenador_id: editavel ? valores.coordenador_id : designacao.coordenador.id,
          portaria: valores.portaria,
          data_inicio: editavel ? valores.data_inicio : designacao.data_inicio,
          data_fim: valores.data_fim,
        },
        designacao.versao
      )
      if (resultado) {
        notificar.sucesso("Designação atualizada com sucesso.")
        onSalvo(resultado)
      }
      return
    }

    const resultado = await criar({
      coordenador_id: valores.coordenador_id,
      portaria: valores.portaria,
      data_inicio: valores.data_inicio,
      data_fim: valores.data_fim,
    })
    if (resultado) {
      notificar.sucesso("Designação registrada com sucesso.")
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
      <FieldGroup>
        {conflito && (
          <Alert variant="destructive" role="alert">
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
        {erro?.code === "DESIGNACAO_SOBREPOSTA" && (
          <Alert variant="destructive" role="alert">
            <AlertDescription>
              Já existe uma designação vigente ou futura para este curso no período informado.
              Encerre a designação atual antes de cadastrar outra, ou ajuste as datas.
            </AlertDescription>
          </Alert>
        )}

        <Field data-invalid={!!erroCampo.coordenador_id}>
          <FieldLabel htmlFor={`${idBase}-coordenador`}>Coordenador</FieldLabel>
          <FieldContent>
            {editavel ? (
              <ComboboxEntidade
                id={`${idBase}-coordenador`}
                itens={itensCombobox}
                value={valores.coordenador_id || null}
                onValueChange={(v) => definirCampo("coordenador_id", v ?? "")}
                carregando={carregandoCandidatos}
                erro={!!erroCandidatos}
                onTentarNovamente={() => buscar("")}
                invalido={!!erroCampo.coordenador_id}
                disabled={salvando}
              />
            ) : (
              <>
                <p className="text-sm">{designacao?.coordenador.nome}</p>
                <p className="text-xs text-muted-foreground">
                  Não editável após o início da vigência — para trocar de coordenador, encerre esta
                  designação e cadastre uma nova.
                </p>
              </>
            )}
            {erroCampo.coordenador_id && <FieldError>{erroCampo.coordenador_id}</FieldError>}
          </FieldContent>
        </Field>

        {tambemEhPI && (
          <Alert role="status" aria-live="polite">
            <InfoIcon aria-hidden="true" />
            <AlertDescription>
              {candidatoSelecionado?.nome} também é Pesquisador(a) Institucional. Ela(e) passará a
              coordenar este curso e continuará podendo avaliar as entregas dele. É permitido, e as
              avaliações feitas por ela(e) neste curso aparecerão marcadas no relatório de
              desempenho.
            </AlertDescription>
          </Alert>
        )}
        {ehAutodesignacao && (
          <Alert role="status" aria-live="polite">
            <InfoIcon aria-hidden="true" />
            <AlertDescription>
              Você está se designando para este curso. A autodesignação é permitida e ficará
              registrada, com a portaria informada.
            </AlertDescription>
          </Alert>
        )}

        <Field data-invalid={!!erroCampo.portaria}>
          <FieldLabel htmlFor={`${idBase}-portaria`}>Portaria</FieldLabel>
          <FieldContent>
            <Input
              id={`${idBase}-portaria`}
              value={valores.portaria}
              onChange={(e) => definirCampo("portaria", e.target.value)}
              disabled={salvando}
            />
            {erroCampo.portaria && <FieldError>{erroCampo.portaria}</FieldError>}
          </FieldContent>
        </Field>

        <Field data-invalid={!!erroCampo.data_inicio}>
          <FieldLabel htmlFor={`${idBase}-inicio`}>Início</FieldLabel>
          <FieldContent>
            {editavel ? (
              <Input
                id={`${idBase}-inicio`}
                type="date"
                value={valores.data_inicio}
                onChange={(e) => definirCampo("data_inicio", e.target.value)}
                disabled={salvando}
              />
            ) : (
              <p className="text-sm">
                {designacao?.data_inicio && formatarDataPura(designacao.data_inicio)}
              </p>
            )}
            {erroCampo.data_inicio && <FieldError>{erroCampo.data_inicio}</FieldError>}
          </FieldContent>
        </Field>

        <Field data-invalid={!!erroCampo.data_fim}>
          <FieldLabel htmlFor={`${idBase}-fim`}>Fim</FieldLabel>
          <FieldContent>
            <Input
              id={`${idBase}-fim`}
              type="date"
              value={valores.data_fim}
              onChange={(e) => definirCampo("data_fim", e.target.value)}
              disabled={salvando}
            />
            {!modoEdicao && (
              <FieldDescription>Deixe o fim em branco para prazo indeterminado.</FieldDescription>
            )}
            {erroCampo.data_fim && <FieldError>{erroCampo.data_fim}</FieldError>}
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
