"use client"

import { InfoIcon } from "lucide-react"
import { useEffect, useId, useState } from "react"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Field, FieldContent, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { formatarDataPura, hojeDataPura } from "@/lib/formato"
import { useAtualizarDesignacao } from "../hooks/use-atualizar-designacao"
import type { Designacao } from "../types"

function diaSeguinte(iso: string): string {
  const [ano, mes, dia] = iso.split("-").map(Number)
  const data = new Date(Date.UTC(ano, mes - 1, dia + 1))
  const isoSeguinte = `${data.getUTCFullYear()}-${String(data.getUTCMonth() + 1).padStart(2, "0")}-${String(data.getUTCDate()).padStart(2, "0")}`
  return formatarDataPura(isoSeguinte)
}

interface EncerrarDesignacaoModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  designacao: Designacao | null
  cursoNome: string
  onSucesso: (item: Designacao) => void
}

export function EncerrarDesignacaoModal({
  open,
  onOpenChange,
  designacao,
  cursoNome,
  onSucesso,
}: EncerrarDesignacaoModalProps) {
  const idBase = useId()
  const { atualizar, isSubmitting: salvando, limparErro } = useAtualizarDesignacao()
  const [novoFim, setNovoFim] = useState(hojeDataPura())
  const [erroCampo, setErroCampo] = useState<string | undefined>()

  useEffect(() => {
    if (open) {
      setNovoFim(hojeDataPura())
      setErroCampo(undefined)
      limparErro()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, designacao])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!designacao) return
    if (designacao.data_inicio && novoFim < designacao.data_inicio) {
      setErroCampo("A data de fim não pode ser anterior à data de início.")
      return
    }
    setErroCampo(undefined)
    const resultado = await atualizar(
      designacao.id,
      {
        coordenador_id: designacao.coordenador.id,
        portaria: designacao.portaria,
        data_inicio: designacao.data_inicio,
        data_fim: novoFim,
      },
      designacao.versao
    )
    if (resultado) {
      notificar.sucesso(
        `Designação de ${designacao.coordenador.nome} encerrada em ${formatarDataPura(novoFim)}.`
      )
      onSucesso(resultado)
      onOpenChange(false)
    }
  }

  const ultimaVigente = designacao?.outras_designacoes_vigentes === 0

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>
            Encerrar a designação de {designacao?.coordenador.nome} em {cursoNome}?
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} aria-busy={salvando}>
          <FieldGroup>
            <Field data-invalid={!!erroCampo}>
              <FieldLabel htmlFor={`${idBase}-fim`}>Nova data de fim</FieldLabel>
              <FieldContent>
                <Input
                  id={`${idBase}-fim`}
                  type="date"
                  value={novoFim}
                  min={designacao?.data_inicio}
                  onChange={(e) => setNovoFim(e.target.value)}
                  disabled={salvando}
                />
                {erroCampo && <FieldError>{erroCampo}</FieldError>}
              </FieldContent>
            </Field>

            <Alert role="status" aria-live="polite">
              <InfoIcon aria-hidden="true" />
              <AlertDescription>
                A partir de {diaSeguinte(novoFim)} o curso fica sem responsável: ninguém poderá
                registrar entrega, e os planos vigentes continuam sendo cobrados.
              </AlertDescription>
            </Alert>

            <Alert role="status" aria-live="polite">
              <InfoIcon aria-hidden="true" />
              <AlertDescription>
                {ultimaVigente
                  ? `${designacao?.coordenador.nome} deixa de ter o perfil de Coordenador de Curso, porque esta é a última designação vigente dele(a). Os demais perfis não mudam.`
                  : `Se esta for a última designação vigente de ${designacao?.coordenador.nome}, ele(a) deixará de ter o perfil de Coordenador de Curso. Os demais perfis não mudam.`}
              </AlertDescription>
            </Alert>

            <Alert role="status">
              <InfoIcon aria-hidden="true" />
              <AlertDescription>
                As entregas já feitas continuam contando para o curso, com o nome de quem as enviou.
              </AlertDescription>
            </Alert>
          </FieldGroup>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={salvando}
            >
              Cancelar
            </Button>
            <LoadingButton type="submit" loading={salvando} loadingText="Encerrando...">
              Encerrar
            </LoadingButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
