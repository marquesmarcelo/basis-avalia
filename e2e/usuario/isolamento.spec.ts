import { test, expect } from "@playwright/test"
import { login } from "../helpers/auth"

test("PI da FSA não alcança usuário de outra instituição pela tela", async ({ page }) => {
  await login(page, { instituicao: "Faculdade Serra Azul (FSA)", email: "maria.souza@fsa.edu.br", senha: "reuniao-nde-2026" })
  await expect(page).toHaveURL("/app")

  await page.goto("/app/usuarios")
  await page.getByLabel("Nome ou e-mail").fill("Renata Coimbra")
  await page.getByRole("button", { name: "Pesquisar" }).click()

  await expect(page.getByText("Nenhum usuário encontrado.")).toBeVisible()
  await expect(page.getByText("Renata Coimbra")).not.toBeVisible()
})
