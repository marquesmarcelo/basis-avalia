import { test as base, type Page, expect } from "@playwright/test"
import { API_URL } from "./env"

export interface NovoUsuario {
  nome: string
  email: string
  perfis: string[]
  senha: string
}

async function preencherECriar(page: Page, dados: NovoUsuario): Promise<string> {
  await page.getByRole("link", { name: "+ Novo" }).click()
  await expect(page).toHaveURL("/app/usuarios/novo")

  await page.getByLabel("Nome").fill(dados.nome)
  await page.getByLabel("E-mail").fill(dados.email)

  for (const perfil of dados.perfis) {
    const caixa = page.getByRole("checkbox", { name: perfil })
    if (!(await caixa.isChecked())) {
      await caixa.click()
    }
  }

  await page.getByLabel("Senha inicial").fill(dados.senha)

  const [resposta] = await Promise.all([
    page.waitForResponse((r) => r.url().includes(`${API_URL}/usuarios`) && r.request().method() === "POST"),
    page.getByRole("button", { name: "Salvar", exact: true }).click(),
  ])
  expect(resposta.ok()).toBeTruthy()
  const corpo = await resposta.json()
  return corpo.id as string
}

export async function abrirUsuariosEPesquisar(page: Page): Promise<void> {
  await page.goto("/app/usuarios")
  await page.getByRole("button", { name: "Pesquisar" }).click()
}

export async function editarNomeViaUI(page: Page, nomeAtual: string, novoNome: string): Promise<void> {
  await page.getByRole("link", { name: `Editar ${nomeAtual}` }).click()
  await page.getByRole("heading", { name: "Editar usuário" }).waitFor()
  await page.getByLabel("Nome").fill(novoNome)
  await page.getByRole("button", { name: "Salvar", exact: true }).click()
}

type Fixtures = {
  criarUsuarioViaUI: (page: Page, dados: NovoUsuario) => Promise<string>
}

// Fixture, não função solta (CLAUDE.md, "quem cria, limpa"): o teardown de
// um fixture do Playwright é sempre aguardado antes do teste terminar — um
// `page.on("close", ...)` não é, o handler é fire-and-forget e o contexto
// do navegador pode fechar antes da requisição de exclusão completar. Era
// a causa do resíduo em usuario/crud.spec.ts quando o teste falhava antes
// do passo de excluir pela UI.
export const test = base.extend<Fixtures>({
  criarUsuarioViaUI: async ({}, use) => {
    const criados: { page: Page; id: string }[] = []
    await use(async (page, dados) => {
      const id = await preencherECriar(page, dados)
      criados.push({ page, id })
      return id
    })
    for (const { page, id } of criados) {
      const resposta = await page.request.delete(`${API_URL}/usuarios/${id}`)
      if (!resposta.ok() && resposta.status() !== 404) {
        // 404 é aceitável — o próprio teste já excluiu pela UI antes de
        // terminar (caminho feliz). Qualquer outro código é resíduo real,
        // e precisa aparecer no relatório, não ser engolido.
        console.error(`teardown de criarUsuarioViaUI: falha ao excluir ${id} — HTTP ${resposta.status()}`)
      }
    }
  },
})
