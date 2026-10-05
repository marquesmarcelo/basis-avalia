"use client"

import { useEffect, useId, useRef, useState } from "react"
import { PasswordInput } from "@/components/shared/forms/password-input"
import { useDirtyState } from "@/components/shared/hooks/use-dirty-state"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { CheckboxGroup } from "@/components/ui/checkbox-group"
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSet,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { focarPrimeiroCampoComErro } from "@/lib/foco"
import { emailValido, obrigatorio } from "@/lib/validacao"
import { useAtualizarUsuario } from "../hooks/use-atualizar-usuario"
import { useCriarUsuario } from "../hooks/use-criar-usuario"
import type { Usuario } from "../types"
import { PERFIS_ESCOLHIVEIS } from "../types"

interface UsuarioFormProps {
  basePath: string
  usuario: Usuario | null
  usuarioAtualId: string
  perfilFixo?: string
  rotuloPerfilFixo?: string
  rotuloSalvar?: string
  rotuloSalvando?: string
  mensagemSucesso?: string
  onSalvo: (usuario: Usuario, criado: boolean) => void
  onCancelar: () => void
  onDirtyChange?: (sujo: boolean) => void
  onConflito?: () => void
  criarOutro?: boolean
}

function ordenarPerfis(lista: string[]): string[] {
  return PERFIS_ESCOLHIVEIS.map((p) => p.value).filter((v) => lista.includes(v))
}

export function UsuarioForm({
  basePath,
  usuario,
  usuarioAtualId,
  perfilFixo,
  rotuloPerfilFixo,
  rotuloSalvar = "Salvar",
  rotuloSalvando = "Salvando...",
  mensagemSucesso,
  onSalvo,
  onCancelar,
  onDirtyChange,
  onConflito,
  criarOutro = false,
}: UsuarioFormProps) {
  const modoEdicao = !!usuario
  const editandoProprioRegistro = modoEdicao && usuario?.id === usuarioAtualId
  const idBase = useId()
  const primeiroCampoRef = useRef<HTMLInputElement>(null)
  const formRef = useRef<HTMLFormElement>(null)

  const perfisIniciais = perfilFixo ? [perfilFixo] : usuario ? usuario.perfis : ["aluno"]

  const { valores, definirCampo, sujo, reiniciar } = useDirtyState({
    nome: usuario?.nome ?? "",
    email: usuario?.email ?? "",
    perfisTexto: ordenarPerfis(perfisIniciais).join(","),
    senha: "",
  })

  const [erroNome, setErroNome] = useState<string | undefined>()
  const [erroEmail, setErroEmail] = useState<string | undefined>()
  const [erroSenha, setErroSenha] = useState<string | undefined>()
  const [conflito, setConflito] = useState(false)
  const [avisoAlunoAutomatico, setAvisoAlunoAutomatico] = useState(false)

  const {
    criar,
    isSubmitting: criando,
    error: erroCriar,
    limparErro: limparErroCriar,
  } = useCriarUsuario(basePath)
  const {
    atualizar,
    isSubmitting: atualizando,
    error: erroAtualizar,
    limparErro: limparErroAtualizar,
  } = useAtualizarUsuario(basePath)

  const salvando = criando || atualizando
  const erro = erroCriar ?? erroAtualizar
  const perfisSelecionados = valores.perfisTexto ? valores.perfisTexto.split(",") : []

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
    setErroEmail(undefined)
    setErroSenha(undefined)
    setConflito(false)
    setAvisoAlunoAutomatico(false)
    limparErroCriar()
    limparErroAtualizar()
    requestAnimationFrame(() => primeiroCampoRef.current?.focus())
  }

  function handlePerfisChange(novoValor: string[]) {
    if (novoValor.length === 0) {
      setAvisoAlunoAutomatico(true)
      definirCampo("perfisTexto", "aluno")
      return
    }
    setAvisoAlunoAutomatico(false)
    definirCampo("perfisTexto", ordenarPerfis(novoValor).join(","))
  }

  async function submeter(continuarCriando: boolean) {
    const semNome = obrigatorio(valores.nome, "O nome é obrigatório.")
    const semEmail =
      obrigatorio(valores.email, "O e-mail é obrigatório.") ??
      emailValido(valores.email, "Informe um e-mail válido.")
    const semSenha = !modoEdicao ? obrigatorio(valores.senha, "A senha é obrigatória.") : undefined
    setErroNome(semNome)
    setErroEmail(semEmail)
    setErroSenha(semSenha)
    if (semNome || semEmail || semSenha) {
      focarPrimeiroCampoComErro(formRef.current)
      return
    }

    const perfis = perfilFixo ? undefined : perfisSelecionados

    if (modoEdicao && usuario) {
      const resultado = await atualizar(
        usuario.id,
        { nome: valores.nome, email: valores.email, perfis },
        usuario.versao
      )
      if (resultado) {
        notificar.sucesso(mensagemSucesso ?? "Usuário atualizado com sucesso.")
        onSalvo(resultado, false)
      }
      return
    }

    const resultado = await criar({
      nome: valores.nome,
      email: valores.email,
      perfis,
      senha: valores.senha,
    })
    if (resultado) {
      notificar.sucesso(mensagemSucesso ?? "Usuário cadastrado com sucesso.")
      if (continuarCriando) {
        limparFormulario()
      } else {
        onSalvo(resultado, true)
      }
    }
  }

  const rotuloPerfil =
    rotuloPerfilFixo ?? PERFIS_ESCOLHIVEIS.find((p) => p.value === perfilFixo)?.label

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

        <Field data-invalid={!!erroEmail || erro?.campo === "email"}>
          <FieldLabel htmlFor={`${idBase}-email`}>E-mail</FieldLabel>
          <FieldContent>
            <Input
              id={`${idBase}-email`}
              type="email"
              value={valores.email}
              onChange={(e) => definirCampo("email", e.target.value)}
              disabled={salvando}
              aria-invalid={!!erroEmail || erro?.campo === "email"}
              aria-describedby={erroEmail ? `${idBase}-email-erro` : undefined}
            />
            {erroEmail && <FieldError id={`${idBase}-email-erro`}>{erroEmail}</FieldError>}
            {!erroEmail && erro?.campo === "email" && <FieldError>{erro.message}</FieldError>}
          </FieldContent>
        </Field>

        {perfilFixo ? (
          <Field className="md:col-span-full">
            <FieldLabel>Perfil</FieldLabel>
            <FieldContent>
              <p className="text-sm">Perfil: {rotuloPerfil}</p>
            </FieldContent>
          </Field>
        ) : (
          <div className="md:col-span-full">
            <CheckboxGroup
              value={perfisSelecionados}
              onValueChange={handlePerfisChange}
              disabled={salvando || editandoProprioRegistro}
            >
              <FieldSet>
                <FieldLegend variant="label">Perfis</FieldLegend>
                {PERFIS_ESCOLHIVEIS.map((p) => (
                  <FieldLabel key={p.value}>
                    <Checkbox value={p.value} />
                    {p.label}
                  </FieldLabel>
                ))}
                <FieldDescription>
                  Pode marcar mais de um. Uma pessoa pode ser aluno e professor ao mesmo tempo.
                </FieldDescription>
                {avisoAlunoAutomatico && (
                  <FieldDescription>
                    Todo usuário precisa de ao menos um perfil. Aluno foi marcado automaticamente.
                  </FieldDescription>
                )}
                {editandoProprioRegistro && (
                  <FieldDescription>
                    Você não pode alterar os seus próprios perfis.
                  </FieldDescription>
                )}
              </FieldSet>
            </CheckboxGroup>
          </div>
        )}

        {!modoEdicao && (
          <Field className="md:col-span-full" data-invalid={!!erroSenha}>
            <FieldLabel htmlFor={`${idBase}-senha`}>Senha inicial</FieldLabel>
            <FieldContent>
              <PasswordInput
                id={`${idBase}-senha`}
                autoComplete="new-password"
                value={valores.senha}
                onChange={(e) => definirCampo("senha", e.target.value)}
                disabled={salvando}
                aria-invalid={!!erroSenha}
                aria-describedby={erroSenha ? `${idBase}-senha-erro` : undefined}
              />
              {erroSenha && <FieldError id={`${idBase}-senha-erro`}>{erroSenha}</FieldError>}
              {perfilFixo && (
                <FieldDescription>
                  A pessoa precisará definir uma senha própria no primeiro acesso. Comunique esta
                  senha a ela.
                </FieldDescription>
              )}
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
            loadingText={rotuloSalvando}
            onClick={() => submeter(true)}
          >
            Salvar e cadastrar outro
          </LoadingButton>
        )}
        <LoadingButton type="submit" loading={salvando} loadingText={rotuloSalvando}>
          {rotuloSalvar}
        </LoadingButton>
      </div>
    </form>
  )
}
