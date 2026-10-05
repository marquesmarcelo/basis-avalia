import { type Page, expect } from "@playwright/test"

export interface Credenciais {
  instituicao?: string
  email: string
  senha: string
}

export async function login(page: Page, credenciais: Credenciais): Promise<void> {
  await page.goto("/")
  await page.getByLabel("Instituição").click()
  if (credenciais.instituicao) {
    await page.getByRole("option", { name: credenciais.instituicao, exact: true }).click()
  } else {
    await page.getByRole("option", { name: "Administração do sistema" }).click()
  }
  await page.getByLabel("E-mail", { exact: true }).fill(credenciais.email)
  await page.getByLabel("Senha", { exact: true }).fill(credenciais.senha)
  await page.getByRole("button", { name: "Entrar" }).click()
}

export async function logout(page: Page): Promise<void> {
  if (await page.getByRole("menu").isVisible().catch(() => false)) {
    await page.keyboard.press("Escape")
    await expect(page.getByRole("menu")).not.toBeVisible()
  }
  await page.getByRole("button", { name: "Abrir menu do usuário" }).click()
  await page.getByRole("menuitem", { name: "Sair" }).click()
  await expect(page).toHaveURL("/")
}
