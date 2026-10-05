"use client"

import { useEffect, useId, useState } from "react"
import { SelectComRotulo } from "@/components/shared/forms/select-com-rotulo"
import { useDirtyState } from "@/components/shared/hooks/use-dirty-state"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Field, FieldContent, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Textarea } from "@/components/ui/textarea"
import { useAtualizarPlano } from "../hooks/use-atualizar-plano"
import type { Plano } from "../types"

interface PlanoDadosFormProps {
  plano: Plano
  somenteLeitura: boolean
  onSalvo: (item: Plano) => void
  onConflito: () => void
}

export function PlanoDadosForm({
  plano,
  somenteLeitura,
  onSalvo,
  onConflito,
}: PlanoDadosFormProps) {
  const idBase = useId()
  const { valores, definirCampo, sujo, reiniciar } = useDirtyState({
    descricao: plano.descricao,
    objetivo_geral: plano.objetivo_geral,
    resultados_esperados: plano.resultados_esperados,
    alinhamento_pdi: plano.alinhamento_pdi,
    alinhamento_ppc: plano.alinhamento_ppc,
    aprovacao_data: plano.aprovacao?.data ?? "",
    aprovacao_orgao: plano.aprovacao?.orgao ?? "",
  })
  const [erros, setErros] = useState<Record<string, string>>({})

  const { atualizar, isSubmitting, error, limparErro } = useAtualizarPlano()

  useEffect(() => {
    reiniciar()
    setErros({})
    limparErro()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [plano.id, plano.versao])

  useEffect(() => {
    if (error?.code === "CONFLITO_DE_VERSAO") onConflito()
    if (error?.code === "APROVACAO_INCOMPLETA") {
      setErros((e) => ({
        ...e,
        aprovacao_data: "Informe a data e o órgão de aprovação, ou deixe os dois em branco.",
      }))
    }
  }, [error, onConflito])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const novosErros: Record<string, string> = {}
    if (!valores.descricao.trim()) novosErros.descricao = "A descrição é obrigatória."
    if (!valores.objetivo_geral.trim())
      novosErros.objetivo_geral = "O objetivo geral é obrigatório."
    if (!valores.resultados_esperados.trim())
      novosErros.resultados_esperados = "Os resultados esperados são obrigatórios."
    setErros(novosErros)
    if (Object.keys(novosErros).length > 0) return

    const resultado = await atualizar(
      plano.id,
      {
        curso_id: plano.curso.id,
        periodo_id: plano.periodo.id,
        descricao: valores.descricao,
        objetivo_geral: valores.objetivo_geral,
        resultados_esperados: valores.resultados_esperados,
        alinhamento_pdi: valores.alinhamento_pdi,
        alinhamento_ppc: valores.alinhamento_ppc,
        aprovacao_data: valores.aprovacao_data,
        aprovacao_orgao: valores.aprovacao_orgao,
      },
      plano.versao
    )
    if (resultado) {
      notificar.sucesso("Plano atualizado.")
      onSalvo(resultado)
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      aria-busy={isSubmitting}
      onKeyDown={(e) => {
        if ((e.ctrlKey || e.metaKey) && e.key === "s") {
          e.preventDefault()
          if (!somenteLeitura) handleSubmit(e)
        }
      }}
    >
      <FieldGroup className="md:grid md:grid-cols-2 md:gap-4">
        {error?.code === "CONFLITO_DE_VERSAO" && (
          <Alert variant="destructive" className="md:col-span-2">
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

        <Field className="md:col-span-full" data-invalid={!!erros.descricao}>
          <FieldLabel htmlFor={`${idBase}-descricao`}>Descrição</FieldLabel>
          <FieldContent>
            <Textarea
              id={`${idBase}-descricao`}
              value={valores.descricao}
              onChange={(e) => definirCampo("descricao", e.target.value)}
              disabled={somenteLeitura || isSubmitting}
            />
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
              disabled={somenteLeitura || isSubmitting}
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
              disabled={somenteLeitura || isSubmitting}
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
              disabled={somenteLeitura || isSubmitting}
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
              disabled={somenteLeitura || isSubmitting}
            />
          </FieldContent>
        </Field>

        <Field data-invalid={!!erros.aprovacao_data}>
          <FieldLabel htmlFor={`${idBase}-aprovacao-data`}>Data de aprovação</FieldLabel>
          <FieldContent>
            <input
              id={`${idBase}-aprovacao-data`}
              type="date"
              className="h-9 w-full rounded-md border px-3 text-sm"
              value={valores.aprovacao_data}
              onChange={(e) => definirCampo("aprovacao_data", e.target.value)}
              disabled={somenteLeitura || isSubmitting}
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
              disabled={somenteLeitura || isSubmitting}
              itens={[
                { value: "nenhum", label: "Ainda não informado" },
                { value: "nde", label: "NDE" },
                { value: "colegiado_curso", label: "Colegiado de curso" },
              ]}
            />
          </FieldContent>
        </Field>
      </FieldGroup>

      {!somenteLeitura && (
        <div className="mt-4 flex justify-end gap-2">
          <Button
            type="button"
            variant="outline"
            disabled={!sujo || isSubmitting}
            onClick={() => reiniciar()}
          >
            Cancelar
          </Button>
          <LoadingButton type="submit" loading={isSubmitting} loadingText="Salvando...">
            Salvar
          </LoadingButton>
        </div>
      )}
    </form>
  )
}
