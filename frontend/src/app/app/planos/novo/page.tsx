"use client"

import { useRouter } from "next/navigation"
import { useEffect, useId, useRef, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { ComboboxEntidade } from "@/components/shared/forms/combobox-entidade"
import { SelectComRotulo } from "@/components/shared/forms/select-com-rotulo"
import { useDirtyState } from "@/components/shared/hooks/use-dirty-state"
import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Button } from "@/components/ui/button"
import { Field, FieldContent, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Textarea } from "@/components/ui/textarea"
import { useCriarPlano } from "@/features/plano/hooks/use-criar-plano"
import { useCursosSugestoes } from "@/features/plano/hooks/use-cursos-sugestoes"
import { usePeriodosSugestoes } from "@/features/plano/hooks/use-periodos-sugestoes"
import { obrigatorio } from "@/lib/validacao"

export default function NovoPlanoPage() {
  const router = useRouter()
  const idBase = useId()
  const primeiroCampoRef = useRef<HTMLButtonElement>(null)

  const [cursoId, setCursoId] = useState<string | null>(null)
  const [periodoId, setPeriodoId] = useState<string | null>(null)
  const {
    valores,
    definirCampo,
    sujo: sujoTexto,
    reiniciar,
  } = useDirtyState({
    descricao: "",
    objetivo_geral: "",
    resultados_esperados: "",
    alinhamento_pdi: "",
    alinhamento_ppc: "",
    aprovacao_data: "",
    aprovacao_orgao: "",
  })
  const sujo = sujoTexto || !!cursoId || !!periodoId

  const [erros, setErros] = useState<Record<string, string>>({})
  const [confirmarDescarte, setConfirmarDescarte] = useState(false)

  const { itens: cursos, isLoading: carregandoCursos, buscar: buscarCursos } = useCursosSugestoes()
  const {
    itens: periodos,
    isLoading: carregandoPeriodos,
    buscar: buscarPeriodos,
  } = usePeriodosSugestoes()
  const { criar, isSubmitting, error, limparErro } = useCriarPlano()

  useEffect(() => {
    buscarCursos("")
    buscarPeriodos("")
    requestAnimationFrame(() => primeiroCampoRef.current?.focus())
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (error?.code === "CURSO_INATIVO")
      setErros((e) => ({ ...e, curso_id: "Curso inativo não pode receber plano." }))
    if (error?.code === "APROVACAO_INCOMPLETA") {
      setErros((e) => ({
        ...e,
        aprovacao_data: "Informe a data e o órgão de aprovação, ou deixe os dois em branco.",
      }))
    }
    if (error?.code === "PLANO_DUPLICADO") {
      setErros((e) => ({
        ...e,
        periodo_id: `Já existe um plano para ${cursos.find((c) => c.id === cursoId)?.nome ?? "este curso"} neste período.`,
      }))
    }
  }, [error, cursoId, cursos])

  function handleCancelar() {
    if (sujo) {
      setConfirmarDescarte(true)
      return
    }
    router.push("/app/planos")
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    limparErro()
    const novosErros: Record<string, string> = {}
    if (!cursoId) novosErros.curso_id = "Selecione um curso."
    if (!periodoId) novosErros.periodo_id = "Selecione um período."
    const semDescricao = obrigatorio(valores.descricao, "A descrição é obrigatória.")
    if (semDescricao) novosErros.descricao = semDescricao
    const semObjetivo = obrigatorio(valores.objetivo_geral, "O objetivo geral é obrigatório.")
    if (semObjetivo) novosErros.objetivo_geral = semObjetivo
    const semResultados = obrigatorio(
      valores.resultados_esperados,
      "Os resultados esperados são obrigatórios."
    )
    if (semResultados) novosErros.resultados_esperados = semResultados
    setErros(novosErros)
    if (Object.keys(novosErros).length > 0 || !cursoId || !periodoId) return

    const resultado = await criar({
      curso_id: cursoId,
      periodo_id: periodoId,
      descricao: valores.descricao,
      objetivo_geral: valores.objetivo_geral,
      resultados_esperados: valores.resultados_esperados,
      alinhamento_pdi: valores.alinhamento_pdi,
      alinhamento_ppc: valores.alinhamento_ppc,
      aprovacao_data: valores.aprovacao_data,
      aprovacao_orgao: valores.aprovacao_orgao,
    })
    if (resultado) {
      notificar.sucesso("Plano criado em rascunho.")
      reiniciar()
      router.push(`/app/planos/${resultado.id}`)
    }
  }

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Planos", href: "/app/planos" },
          { rotulo: "Novo" },
        ]}
      />

      <h1 className="text-2xl font-semibold">Novo plano de ação</h1>

      <form
        onSubmit={handleSubmit}
        aria-busy={isSubmitting}
        className="max-w-4xl"
        onKeyDown={(e) => {
          if ((e.ctrlKey || e.metaKey) && e.key === "s") {
            e.preventDefault()
            handleSubmit(e)
          }
        }}
      >
        <FieldGroup className="md:grid md:grid-cols-2 md:gap-4">
          <Field data-invalid={!!erros.curso_id}>
            <FieldLabel htmlFor={`${idBase}-curso`}>Curso</FieldLabel>
            <FieldContent>
              <ComboboxEntidade
                id={`${idBase}-curso`}
                itens={cursos.map((c) => ({ value: c.id, label: c.nome }))}
                value={cursoId}
                onValueChange={setCursoId}
                carregando={carregandoCursos}
                onTentarNovamente={() => buscarCursos("")}
                disabled={isSubmitting}
              />
              {erros.curso_id && <FieldError>{erros.curso_id}</FieldError>}
            </FieldContent>
          </Field>

          <Field data-invalid={!!erros.periodo_id}>
            <FieldLabel htmlFor={`${idBase}-periodo`}>Período</FieldLabel>
            <FieldContent>
              <ComboboxEntidade
                id={`${idBase}-periodo`}
                itens={periodos.map((p) => ({ value: p.id, label: p.nome }))}
                value={periodoId}
                onValueChange={setPeriodoId}
                carregando={carregandoPeriodos}
                onTentarNovamente={() => buscarPeriodos("")}
                disabled={isSubmitting}
              />
              {erros.periodo_id && <FieldError>{erros.periodo_id}</FieldError>}
            </FieldContent>
          </Field>

          <Field className="md:col-span-full" data-invalid={!!erros.descricao}>
            <FieldLabel htmlFor={`${idBase}-descricao`}>Descrição</FieldLabel>
            <FieldContent>
              <Textarea
                id={`${idBase}-descricao`}
                value={valores.descricao}
                onChange={(e) => definirCampo("descricao", e.target.value)}
                disabled={isSubmitting}
              />
              <p className="text-sm text-muted-foreground">
                Não inclua nome de pessoa neste campo — ele vai para um documento que circula.
              </p>
              {erros.descricao && <FieldError>{erros.descricao}</FieldError>}
            </FieldContent>
          </Field>

          <Field className="md:col-span-full" data-invalid={!!erros.objetivo_geral}>
            <FieldLabel htmlFor={`${idBase}-objetivo`}>Objetivo geral</FieldLabel>
            <FieldContent>
              <Textarea
                id={`${idBase}-objetivo`}
                value={valores.objetivo_geral}
                onChange={(e) => definirCampo("objetivo_geral", e.target.value)}
                disabled={isSubmitting}
              />
              {erros.objetivo_geral && <FieldError>{erros.objetivo_geral}</FieldError>}
            </FieldContent>
          </Field>

          <Field className="md:col-span-full" data-invalid={!!erros.resultados_esperados}>
            <FieldLabel htmlFor={`${idBase}-resultados`}>Resultados esperados</FieldLabel>
            <FieldContent>
              <Textarea
                id={`${idBase}-resultados`}
                value={valores.resultados_esperados}
                onChange={(e) => definirCampo("resultados_esperados", e.target.value)}
                disabled={isSubmitting}
              />
              {erros.resultados_esperados && <FieldError>{erros.resultados_esperados}</FieldError>}
            </FieldContent>
          </Field>

          <Field>
            <FieldLabel htmlFor={`${idBase}-pdi`}>Alinhamento com o PDI</FieldLabel>
            <FieldContent>
              <Textarea
                id={`${idBase}-pdi`}
                value={valores.alinhamento_pdi}
                onChange={(e) => definirCampo("alinhamento_pdi", e.target.value)}
                disabled={isSubmitting}
              />
            </FieldContent>
          </Field>

          <Field>
            <FieldLabel htmlFor={`${idBase}-ppc`}>Alinhamento com o PPC</FieldLabel>
            <FieldContent>
              <Textarea
                id={`${idBase}-ppc`}
                value={valores.alinhamento_ppc}
                onChange={(e) => definirCampo("alinhamento_ppc", e.target.value)}
                disabled={isSubmitting}
              />
            </FieldContent>
          </Field>

          <Field data-invalid={!!erros.aprovacao_data}>
            <FieldLabel htmlFor={`${idBase}-aprovacao-data`}>
              Data de aprovação (opcional)
            </FieldLabel>
            <FieldContent>
              <input
                id={`${idBase}-aprovacao-data`}
                type="date"
                className="h-9 w-full rounded-md border px-3 text-sm"
                value={valores.aprovacao_data}
                onChange={(e) => definirCampo("aprovacao_data", e.target.value)}
                disabled={isSubmitting}
              />
              {erros.aprovacao_data && <FieldError>{erros.aprovacao_data}</FieldError>}
            </FieldContent>
          </Field>

          <Field>
            <FieldLabel htmlFor={`${idBase}-aprovacao-orgao`}>Órgão de aprovação</FieldLabel>
            <FieldContent>
              <SelectComRotulo
                id={`${idBase}-aprovacao-orgao`}
                value={valores.aprovacao_orgao || "nenhum"}
                onValueChange={(v) => definirCampo("aprovacao_orgao", v === "nenhum" ? "" : v)}
                disabled={isSubmitting}
                itens={[
                  { value: "nenhum", label: "Ainda não informado" },
                  { value: "nde", label: "NDE" },
                  { value: "colegiado_curso", label: "Colegiado de curso" },
                ]}
              />
            </FieldContent>
          </Field>
        </FieldGroup>

        <div className="mt-6 flex justify-end gap-2">
          <Button type="button" variant="outline" onClick={handleCancelar} disabled={isSubmitting}>
            Cancelar
          </Button>
          <LoadingButton type="submit" loading={isSubmitting} loadingText="Criando...">
            Criar plano
          </LoadingButton>
        </div>
      </form>

      <ConfirmDialog
        open={confirmarDescarte}
        onOpenChange={setConfirmarDescarte}
        titulo="Descartar alterações?"
        descricao="Você tem alterações não salvas. Deseja descartá-las?"
        rotuloConfirmar="Descartar alterações"
        rotuloCancelar="Continuar editando"
        destrutivo
        onConfirmar={() => {
          setConfirmarDescarte(false)
          router.push("/app/planos")
        }}
      />
    </div>
  )
}
