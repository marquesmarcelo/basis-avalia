import type { Periodo } from "../types"

export const ROTULO_SITUACAO: Record<Periodo["situacao"], string> = {
  nao_iniciado: "Não iniciado",
  aberto: "Aberto",
  encerrado: "Encerrado",
}

// calcularSituacao espelha Periodo.SituacaoEm do backend (periodo.go):
// hoje antes do início -> não iniciado; hoje depois do fim -> encerrado;
// senão, aberto. Comparação por string funciona porque as três datas
// chegam em ISO "AAAA-MM-DD" (zero-padded), que ordena igual à ordem
// cronológica. "hoje" precisa vir do fuso de exibição (hojeDataPura em
// lib/formato.ts), nunca de new Date() puro — senão o resultado muda
// perto da virada do dia dependendo do fuso do navegador.
export function calcularSituacao(
  dataInicio: string,
  dataFim: string,
  hojeISO: string
): Periodo["situacao"] {
  if (hojeISO < dataInicio) return "nao_iniciado"
  if (hojeISO > dataFim) return "encerrado"
  return "aberto"
}
