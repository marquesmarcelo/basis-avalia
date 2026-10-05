"use client"

type Ouvinte = (ativo: boolean) => void

let contador = 0
const ouvintes = new Set<Ouvinte>()

function notificar() {
  const ativo = contador > 0
  ouvintes.forEach((fn) => {
    fn(ativo)
  })
}

export function iniciarCarregamento() {
  contador += 1
  notificar()
}

export function concluirCarregamento() {
  contador = Math.max(0, contador - 1)
  notificar()
}

export function inscreverCarregamento(fn: Ouvinte): () => void {
  ouvintes.add(fn)
  return () => {
    ouvintes.delete(fn)
  }
}
