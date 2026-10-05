"use client"

import { useEffect, useId, useRef, useState } from "react"
import { SelectComRotulo } from "@/components/shared/forms/select-com-rotulo"
import { useDirtyState } from "@/components/shared/hooks/use-dirty-state"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Field, FieldContent, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { focarPrimeiroCampoComErro } from "@/lib/foco"
import { obrigatorio } from "@/lib/validacao"
import { useAtualizarCurso } from "../hooks/use-atualizar-curso"
import { useCriarCurso } from "../hooks/use-criar-curso"
import type { Curso, GrauDeCurso, ModalidadeDeCurso, SituacaoCurso } from "../types"

interface CursoFormProps {
  curso: Curso | null
  onSalvo: (item: Curso) => void
  onCancelar: () => void
  onDirtyChange?: (sujo: boolean) => void
  onConflito?: () => void
  criarOutro?: boolean
}

export function CursoForm({
  curso,
  onSalvo,
  onCancelar,
  onDirtyChange,
  onConflito,
  criarOutro = false,
}: CursoFormProps) {
  const modoEdicao = !!curso
  const idBase = useId()
  const primeiroCampoRef = useRef<HTMLInputElement>(null)
  const formRef = useRef<HTMLFormElement>(null)

  const {
    valores,
    definirCampo,
    sujo: sujoTexto,
    reiniciar,
  } = useDirtyState({
    nome: curso?.nome ?? "",
    codigo_emec: curso?.codigo_emec ?? "",
  })
  const [grau, setGrau] = useState<GrauDeCurso>(curso?.grau ?? "bacharelado")
  const [modalidade, setModalidade] = useState<ModalidadeDeCurso>(curso?.modalidade ?? "presencial")
  const [situacao, setSituacao] = useState<SituacaoCurso>(curso?.situacao ?? "ativo")

  const [erroNome, setErroNome] = useState<string | undefined>()
  const [conflito, setConflito] = useState(false)

  const {
    criar,
    isSubmitting: criando,
    error: erroCriar,
    limparErro: limparErroCriar,
  } = useCriarCurso()
  const {
    atualizar,
    isSubmitting: atualizando,
    error: erroAtualizar,
    limparErro: limparErroAtualizar,
  } = useAtualizarCurso()

  const salvando = criando || atualizando
  const erro = erroCriar ?? erroAtualizar

  const sujo =
    sujoTexto ||
    grau !== (curso?.grau ?? "bacharelado") ||
    modalidade !== (curso?.modalidade ?? "presencial") ||
    situacao !== (curso?.situacao ?? "ativo")

  useEffect(() => {
    onDirtyChange?.(sujo)
  }, [sujo, onDirtyChange])

  useEffect(() => {
    requestAnimationFrame(() => primeiroCampoRef.current?.focus())
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (erro?.code === "CONFLITO_DE_VERSAO") setConflito(true)
    if (erro?.code === "NOME_CURSO_DUPLICADO") {
      setErroNome("Já existe um curso com este nome nesta instituição.")
    }
  }, [erro])

  function limparFormulario() {
    reiniciar()
    setGrau("bacharelado")
    setModalidade("presencial")
    setSituacao("ativo")
    setErroNome(undefined)
    setConflito(false)
    limparErroCriar()
    limparErroAtualizar()
    requestAnimationFrame(() => primeiroCampoRef.current?.focus())
  }

  async function submeter(continuarCriando: boolean) {
    const semNome = obrigatorio(valores.nome, "O nome é obrigatório.")
    setErroNome(semNome)
    if (semNome) {
      focarPrimeiroCampoComErro(formRef.current)
      return
    }

    const input = { nome: valores.nome, codigo_emec: valores.codigo_emec, grau, modalidade }

    if (modoEdicao && curso) {
      const resultado = await atualizar(curso.id, input, curso.versao)
      if (resultado) {
        notificar.sucesso("Curso atualizado com sucesso.")
        onSalvo(resultado)
      }
      return
    }

    const resultado = await criar(input, situacao)
    if (resultado) {
      notificar.sucesso("Curso cadastrado com sucesso.")
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

        <Field className="md:col-span-full" data-invalid={!!erroNome}>
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

        <Field className="md:col-span-full">
          <FieldLabel htmlFor={`${idBase}-codigo`}>Código e-MEC</FieldLabel>
          <FieldContent>
            <Input
              id={`${idBase}-codigo`}
              value={valores.codigo_emec}
              onChange={(e) => definirCampo("codigo_emec", e.target.value)}
              disabled={salvando}
            />
            {erro?.code === "CODIGO_EMEC_CURSO_DUPLICADO" && (
              <FieldError>Já existe um curso com este código e-MEC nesta instituição.</FieldError>
            )}
          </FieldContent>
        </Field>

        <Field>
          <FieldLabel htmlFor={`${idBase}-grau`}>Grau</FieldLabel>
          <FieldContent>
            <SelectComRotulo
              id={`${idBase}-grau`}
              value={grau}
              onValueChange={(v) => setGrau(v as GrauDeCurso)}
              disabled={salvando}
              itens={[
                { value: "bacharelado", label: "Bacharelado" },
                { value: "licenciatura", label: "Licenciatura" },
                { value: "tecnologo", label: "Tecnólogo" },
              ]}
            />
          </FieldContent>
        </Field>

        <Field>
          <FieldLabel htmlFor={`${idBase}-modalidade`}>Modalidade</FieldLabel>
          <FieldContent>
            <SelectComRotulo
              id={`${idBase}-modalidade`}
              value={modalidade}
              onValueChange={(v) => setModalidade(v as ModalidadeDeCurso)}
              disabled={salvando}
              itens={[
                { value: "presencial", label: "Presencial" },
                { value: "a_distancia", label: "A distância" },
              ]}
            />
          </FieldContent>
        </Field>

        {!modoEdicao && (
          <Field className="md:col-span-full">
            <FieldLabel>Situação</FieldLabel>
            <FieldContent>
              <div className="flex items-center gap-4 pt-2">
                <label className="flex items-center gap-2 text-sm">
                  <input
                    type="radio"
                    name={`${idBase}-situacao`}
                    checked={situacao === "ativo"}
                    onChange={() => setSituacao("ativo")}
                    disabled={salvando}
                  />
                  Ativo
                </label>
                <label className="flex items-center gap-2 text-sm">
                  <input
                    type="radio"
                    name={`${idBase}-situacao`}
                    checked={situacao === "inativo"}
                    onChange={() => setSituacao("inativo")}
                    disabled={salvando}
                  />
                  Inativo
                </label>
              </div>
            </FieldContent>
          </Field>
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
