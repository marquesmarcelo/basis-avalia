export function obrigatorio(
  valor: string | null | undefined,
  mensagem: string
): string | undefined {
  if (!valor || valor.trim() === "") {
    return mensagem
  }
  return undefined
}

export function emailValido(valor: string, mensagem: string): string | undefined {
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(valor)) {
    return mensagem
  }
  return undefined
}

export function senhasCoincidem(
  senha: string,
  confirmacao: string,
  mensagem: string
): string | undefined {
  if (senha !== confirmacao) {
    return mensagem
  }
  return undefined
}
