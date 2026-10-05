import { test, expect } from "@playwright/test"
import { login } from "../helpers/auth"
import { API_URL } from "../helpers/env"

// A restauração roda no afterEach, sempre — nunca no corpo do teste. Um
// self-service (`POST /auth/senha`) grava `senha_provisoria=false`
// (autenticacao_repository.go, DefinirSenhaPropria); só a redefinição pelo
// PI (`POST /usuarios/{id}/senha`) grava `senha_provisoria=true` de volta
// — é o único caminho que devolve o cenário de "primeiro acesso" ao seed.
// Cada etapa afirma sucesso em vez de engolir a falha: um afterEach que
// falha em silêncio é exatamente o que deixa o próximo `test.run` herdar
// um usuário fora do estado que o teste presume.
test.afterEach(async ({ page }) => {
  const instituicoes = await page.request.get(`${API_URL}/publico/instituicoes`)
  expect(instituicoes.ok(), "restauração: listar instituições públicas").toBeTruthy()
  const fsaID = (await instituicoes.json()).find((i: { sigla: string }) => i.sigla === "FSA")?.id

  const loginPI = await page.request.post(`${API_URL}/auth/login`, {
    data: { instituicao_id: fsaID, email: "maria.souza@fsa.edu.br", senha: "reuniao-nde-2026" },
  })
  expect(loginPI.ok(), "restauração: login do PI").toBeTruthy()

  const usuarios = await page.request.get(`${API_URL}/usuarios?busca=joao.ribeiro@ies.edu.br`)
  expect(usuarios.ok(), "restauração: buscar João Ribeiro").toBeTruthy()
  const joaoID = (await usuarios.json()).data?.[0]?.id
  expect(joaoID, "restauração: João Ribeiro (FSA) não encontrado").toBeTruthy()

  const redefinida = await page.request.post(`${API_URL}/usuarios/${joaoID}/senha`, {
    data: { senha_nova: "senha-fsa-2026" },
  })
  expect(redefinida.ok(), "restauração: redefinir senha de João Ribeiro").toBeTruthy()

  await page.request.post(`${API_URL}/auth/logout`)
})

test("primeiro acesso com senha provisória força a troca e depois segue para a aplicação", async ({ page }) => {
  await login(page, { instituicao: "Faculdade Serra Azul (FSA)", email: "joao.ribeiro@ies.edu.br", senha: "senha-fsa-2026" })

  await expect(page).toHaveURL("/app/alterar-senha")
  await expect(page.getByRole("heading", { name: "Defina sua senha" })).toBeVisible()

  const novaSenha = "nova-senha-joao-e2e-2026"
  await page.getByLabel("Senha atual", { exact: true }).fill("senha-fsa-2026")
  await page.getByLabel("Nova senha", { exact: true }).fill(novaSenha)
  await page.getByLabel("Confirme a nova senha", { exact: true }).fill(novaSenha)
  await page.getByRole("button", { name: "Salvar" }).click()

  await expect(page).toHaveURL("/app")
})
