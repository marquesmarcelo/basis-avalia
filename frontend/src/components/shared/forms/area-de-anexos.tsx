"use client"

import { FileIcon, Loader2Icon, XIcon } from "lucide-react"
import { useId, useRef, useState } from "react"
import { ProgressoDeEnvio } from "@/components/shared/ui/progresso-de-envio"
import { Button } from "@/components/ui/button"

function formatarTamanho(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

interface ExistenteAnexo {
  id: string
  nome_original: string
  tamanho_bytes: number
}

interface AreaDeAnexosProps {
  tiposAceitos?: string
  limiteBytesPorArquivo?: number
  limiteArquivos?: number
  limiteBytesTotal?: number

  modo: "agregado" | "porArquivo"
  arquivosNaFila?: File[]
  onArquivosNaFilaChange?: (arquivos: File[]) => void
  enviandoAgregado?: boolean

  anexosExistentes?: ExistenteAnexo[]
  onAdicionarArquivo?: (arquivo: File) => Promise<boolean>
  onRemoverAnexo?: (anexoId: string) => Promise<boolean>
  removendoAnexoId?: string | null

  disabled?: boolean
}

export function AreaDeAnexos({
  tiposAceitos = ".pdf,.jpg,.jpeg,.png,.docx,.odt",
  limiteBytesPorArquivo = 10 * 1024 * 1024,
  limiteArquivos = 10,
  limiteBytesTotal = 50 * 1024 * 1024,
  modo,
  arquivosNaFila = [],
  onArquivosNaFilaChange,
  enviandoAgregado = false,
  anexosExistentes = [],
  onAdicionarArquivo,
  onRemoverAnexo,
  removendoAnexoId,
  disabled = false,
}: AreaDeAnexosProps) {
  const inputId = useId()
  const inputRef = useRef<HTMLInputElement>(null)
  const [erroDeConjunto, setErroDeConjunto] = useState<string | null>(null)
  const [errosPorArquivo, setErrosPorArquivo] = useState<Record<string, string>>({})
  const [enviandoIndividual, setEnviandoIndividual] = useState<string | null>(null)

  const extensoesAceitas = tiposAceitos.split(",").map((e) => e.trim().toLowerCase())

  function validarClientSide(arquivo: File): string | null {
    const extensao = "." + (arquivo.name.split(".").pop() ?? "").toLowerCase()
    if (!extensoesAceitas.includes(extensao)) {
      return `«${arquivo.name}»: tipo de arquivo não permitido. Aceitos: PDF, JPG, PNG, DOCX, ODT.`
    }
    if (arquivo.size > limiteBytesPorArquivo) {
      return `«${arquivo.name}»: ${formatarTamanho(arquivo.size)} excede o limite de 10 MB por arquivo.`
    }
    return null
  }

  async function tratarSelecao(lista: FileList | null) {
    if (!lista || lista.length === 0) return
    const selecionados = Array.from(lista)
    setErroDeConjunto(null)

    if (modo === "agregado") {
      const novos: File[] = []
      const novosErros: Record<string, string> = {}
      let somaAtual = arquivosNaFila.reduce((soma, a) => soma + a.size, 0)
      let quantidadeAtual = arquivosNaFila.length

      for (const arquivo of selecionados) {
        const erro = validarClientSide(arquivo)
        if (erro) {
          novosErros[arquivo.name] = erro
          continue
        }
        if (quantidadeAtual + 1 > limiteArquivos || somaAtual + arquivo.size > limiteBytesTotal) {
          setErroDeConjunto(
            "Limite de 10 arquivos ou 50 MB por entrega atingido. Remova algum arquivo para adicionar outro."
          )
          break
        }
        novos.push(arquivo)
        somaAtual += arquivo.size
        quantidadeAtual += 1
      }
      setErrosPorArquivo(novosErros)
      onArquivosNaFilaChange?.([...arquivosNaFila, ...novos])
    } else {
      for (const arquivo of selecionados) {
        const erro = validarClientSide(arquivo)
        if (erro) {
          setErrosPorArquivo((prev) => ({ ...prev, [arquivo.name]: erro }))
          continue
        }
        setEnviandoIndividual(arquivo.name)
        await onAdicionarArquivo?.(arquivo)
        setEnviandoIndividual(null)
      }
    }
    if (inputRef.current) inputRef.current.value = ""
  }

  function removerDaFila(nome: string) {
    onArquivosNaFilaChange?.(arquivosNaFila.filter((a) => a.name !== nome))
  }

  return (
    <div className="space-y-3">
      <div>
        <input
          ref={inputRef}
          id={inputId}
          type="file"
          multiple
          accept={tiposAceitos}
          className="sr-only"
          disabled={disabled}
          onChange={(e) => tratarSelecao(e.target.files)}
        />
        <Button
          type="button"
          variant="outline"
          disabled={disabled}
          aria-label="Selecionar arquivos do computador"
          onClick={() => inputRef.current?.click()}
        >
          + Adicionar arquivos
        </Button>
        <p className="mt-1 text-xs text-muted-foreground">
          A extensão só filtra ruído óbvio. A autoridade é o conteúdo do arquivo, verificado no
          servidor antes de gravar qualquer byte.
        </p>
      </div>

      {erroDeConjunto && (
        <p role="alert" className="text-sm text-destructive">
          {erroDeConjunto}
        </p>
      )}
      {Object.values(errosPorArquivo).map((msg) => (
        <p key={msg} role="alert" className="text-sm text-destructive">
          {msg}
        </p>
      ))}

      {modo === "agregado" && (
        <ul className="space-y-2">
          {arquivosNaFila.map((arquivo) => (
            <li
              key={arquivo.name}
              className="flex items-center gap-2 rounded-md border px-3 py-2 text-sm"
            >
              <FileIcon className="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
              <span className="flex-1 truncate">{arquivo.name}</span>
              <span className="text-muted-foreground">{formatarTamanho(arquivo.size)}</span>
              {!enviandoAgregado && (
                <button
                  type="button"
                  aria-label={`Remover ${arquivo.name}`}
                  onClick={() => removerDaFila(arquivo.name)}
                  className="text-muted-foreground hover:text-destructive"
                >
                  <XIcon className="size-4" aria-hidden="true" />
                </button>
              )}
            </li>
          ))}
          {enviandoAgregado && (
            <li aria-live="polite">
              <ProgressoDeEnvio percentual={0} indeterminado />
            </li>
          )}
        </ul>
      )}

      {modo === "porArquivo" && (
        <ul className="space-y-2">
          {anexosExistentes.map((anexo) => (
            <li
              key={anexo.id}
              className="flex items-center gap-2 rounded-md border px-3 py-2 text-sm"
            >
              <FileIcon className="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
              <span className="flex-1 truncate">{anexo.nome_original}</span>
              <span className="text-muted-foreground">{formatarTamanho(anexo.tamanho_bytes)}</span>
              <button
                type="button"
                aria-label={`Remover ${anexo.nome_original}`}
                disabled={removendoAnexoId === anexo.id}
                onClick={() => onRemoverAnexo?.(anexo.id)}
                className="text-muted-foreground hover:text-destructive disabled:opacity-50"
              >
                {removendoAnexoId === anexo.id ? (
                  <Loader2Icon className="size-4 animate-spin" aria-hidden="true" />
                ) : (
                  <XIcon className="size-4" aria-hidden="true" />
                )}
              </button>
            </li>
          ))}
          {enviandoIndividual && (
            <li
              className="flex items-center gap-2 rounded-md border px-3 py-2 text-sm"
              aria-live="polite"
            >
              <Loader2Icon className="size-4 shrink-0 animate-spin" aria-hidden="true" />
              <span className="flex-1 truncate">{enviandoIndividual}</span>
              <span className="text-muted-foreground">Enviando...</span>
            </li>
          )}
        </ul>
      )}
    </div>
  )
}
