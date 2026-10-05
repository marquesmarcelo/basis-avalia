import { test, expect } from "@playwright/test"
import { login } from "../helpers/auth"

test("menus, popovers e diálogos do shell abrem sem quebrar a aplicação", async ({ page }) => {
  const erros: string[] = []
  page.on("pageerror", (err) => erros.push(err.message))

  await login(page, { instituicao: "Faculdade Serra Azul (FSA)", email: "maria.souza@fsa.edu.br", senha: "reuniao-nde-2026" })
  await expect(page).toHaveURL("/app")

  await page.getByRole("button", { name: "Abrir menu do usuário" }).click()
  await expect(page.getByRole("menu")).toBeVisible()
  await expect(page.getByRole("menuitem", { name: "Sair" })).toBeVisible()
  await page.keyboard.press("Escape")

  await page.keyboard.press("?")
  await expect(page.getByRole("dialog", { name: "Atalhos de teclado" })).toBeVisible()
  await page.keyboard.press("Escape")

  expect(erros, `erros de página capturados: ${erros.join(" | ")}`).toEqual([])
})
