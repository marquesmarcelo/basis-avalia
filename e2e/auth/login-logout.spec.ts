import { test, expect } from "@playwright/test"
import { login, logout } from "../helpers/auth"

test.describe("Login e logout", () => {
  test("Administração do sistema entra pelo item fixo do combo e sai corretamente", async ({ page }) => {
    await login(page, { email: "rafael.toledo@basis-avalia.local", senha: "plataforma-2026" })

    // A conta da plataforma nasce sempre com senha provisória (3.4) — o
    // primeiro login força a troca antes de qualquer outra coisa. Mas a
    // flag não é restaurável sem um terceiro administrador permanente
    // (E-08 impede autoexclusão, E-12 impede autorredefinição — com só
    // dois administradores o ciclo de limpeza nunca fecha, ver metas/
    // catalogo-comum.spec.ts). Na primeira execução da suíte contra um
    // banco recém-semeado ela é sempre verdadeira; nas seguintes, algum
    // teste já consumiu a senha provisória de Rafael — o teste aceita os
    // dois estados, porque o que importa aqui é o item fixo do combo e o
    // logout, não a senha provisória (que já tem dono dedicado em auth/
    // primeiro-acesso.spec.ts, com um usuário comum e restauração real).
    let aindaProvisoria = true
    try {
      await expect(page).toHaveURL("/app/alterar-senha", { timeout: 3000 })
    } catch {
      aindaProvisoria = false
    }
    if (aindaProvisoria) {
      await expect(page.getByRole("heading", { name: "Defina sua senha" })).toBeVisible()
    } else {
      await expect(page).toHaveURL("/app")
    }

    await logout(page)
    await expect(page.getByLabel("Instituição")).toBeVisible()
  })

  test("PI de uma instituição escolhe a instituição no combo, entra e sai", async ({ page }) => {
    await login(page, { instituicao: "Faculdade Serra Azul (FSA)", email: "maria.souza@fsa.edu.br", senha: "reuniao-nde-2026" })

    await expect(page).toHaveURL("/app")
    await expect(page.getByText("Faculdade Serra Azul").first()).toBeVisible()

    await logout(page)
  })

  test("depois de sair, o botão voltar do navegador não mostra tela autenticada", async ({ page }) => {
    await login(page, { instituicao: "Faculdade Serra Azul (FSA)", email: "maria.souza@fsa.edu.br", senha: "reuniao-nde-2026" })
    await expect(page).toHaveURL("/app")

    await page.goto("/app/usuarios")
    await logout(page)

    await page.goBack()
    await expect(page).toHaveURL("/")
    await expect(page.getByLabel("Instituição")).toBeVisible()
    await expect(page.getByRole("heading", { name: /Bem-vindo/ })).not.toBeVisible()
  })
})
