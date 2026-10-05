import { test, expect } from "@playwright/test"
import { login, logout } from "../helpers/auth"

test.describe("Mesmo e-mail em duas instituições", () => {
  test("joao.ribeiro@ies.edu.br autentica na FSA com a senha da FSA", async ({ page }) => {
    await login(page, { instituicao: "Faculdade Serra Azul (FSA)", email: "joao.ribeiro@ies.edu.br", senha: "senha-fsa-2026" })
    await expect(page).toHaveURL(/\/app(\/alterar-senha)?$/)
    await expect(page.getByText("Instituição, e-mail ou senha inválidos.")).not.toBeVisible()
  })

  test("joao.ribeiro@ies.edu.br entra na conta de coordenador do IVV com a senha do IVV", async ({ page }) => {
    await login(page, { instituicao: "Instituto Vale Verde (IVV)", email: "joao.ribeiro@ies.edu.br", senha: "senha-ivv-2026" })

    await expect(page).toHaveURL("/app")
    await expect(page.getByText("Instituto Vale Verde").first()).toBeVisible()

    await page.getByRole("button", { name: "Abrir menu do usuário" }).click()
    await expect(page.getByRole("menu").getByText("Coordenador de Curso").first()).toBeVisible()

    await logout(page)
  })

  test("a senha de uma conta não abre a outra conta do mesmo e-mail", async ({ page }) => {
    await login(page, { instituicao: "Instituto Vale Verde (IVV)", email: "joao.ribeiro@ies.edu.br", senha: "senha-fsa-2026" })

    await expect(page.getByText("Instituição, e-mail ou senha inválidos.")).toBeVisible()
    await expect(page).toHaveURL("/")
  })
})
