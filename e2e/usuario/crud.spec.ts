import { expect } from "@playwright/test"
import { login } from "../helpers/auth"
import { abrirUsuariosEPesquisar, editarNomeViaUI, test } from "../helpers/usuario"

test("PI cria usuário com múltiplos perfis, edita e exclui", async ({ page, criarUsuarioViaUI }) => {
  await login(page, { instituicao: "Faculdade Serra Azul (FSA)", email: "maria.souza@fsa.edu.br", senha: "reuniao-nde-2026" })
  await expect(page).toHaveURL("/app")

  await abrirUsuariosEPesquisar(page)

  const email = `e2e.crud.${Date.now()}@fsa.edu.br`
  const nome = "Usuário E2E Crud"
  await criarUsuarioViaUI(page, { nome, email, perfis: ["Aluno", "Professor"], senha: "primeiro-acesso-e2e-2026" })

  await expect(page.getByText("Usuário cadastrado com sucesso.")).toBeVisible()
  await page.getByRole("button", { name: "Pesquisar" }).click()
  const linha = page.getByRole("row", { name: new RegExp(nome) })
  await expect(linha).toBeVisible()
  await expect(linha).toContainText("Aluno")
  await expect(linha).toContainText("Professor")

  const nomeEditado = "Usuário E2E Crud Editado"
  await editarNomeViaUI(page, nome, nomeEditado)
  await expect(page.getByText("Usuário atualizado com sucesso.")).toBeVisible()
  await page.getByRole("button", { name: "Pesquisar" }).click()
  await expect(page.getByRole("row", { name: new RegExp(nomeEditado) })).toBeVisible()

  await page.getByRole("button", { name: `Excluir ${nomeEditado}` }).click()
  await page.getByRole("alertdialog").getByRole("button", { name: "Excluir", exact: true }).click()
  await expect(page.getByText(`${nomeEditado} foi excluído(a).`)).toBeVisible()
  await expect(page.getByRole("row", { name: new RegExp(nomeEditado) })).not.toBeVisible()
})
