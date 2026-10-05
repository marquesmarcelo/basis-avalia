import { test, expect } from "@playwright/test"
import { login, logout } from "../helpers/auth"
import { API_URL } from "../helpers/env"

// Fluxo crítico de specs/indicadores/design.md §10.3: o isolamento do
// catálogo comum é a primeira informação compartilhada do sistema, e o
// único cenário em que um erro produz vazamento entre clientes.

test.describe("Isolamento do catálogo comum de indicadores", () => {
  test("PI da FSA vê o catálogo comum e o próprio, sem ação na linha do INEP; PI do IVV não vê o próprio da FSA", async ({
    page,
  }) => {
    await login(page, {
      instituicao: "Faculdade Serra Azul (FSA)",
      email: "maria.souza@fsa.edu.br",
      senha: "reuniao-nde-2026",
    })
    await expect(page).toHaveURL("/app")

    await page.goto("/app/indicadores")
    await page.getByRole("button", { name: "Pesquisar" }).click()

    const linhaDoInep = page.getByRole("row", { name: /Núcleo Docente Estruturante/ })
    await expect(linhaDoInep).toBeVisible()
    await expect(linhaDoInep).toContainText("Do INEP")
    await expect(linhaDoInep).toContainText("Somente leitura")
    // A regra é "sem ação de EDIÇÃO nesta linha" — não "sem elemento
    // interativo nenhum" (a linha tem, de propósito, um link de navegação
    // só-leitura para "ver metas"). Checar por nome de ação em vez de por
    // tag/role (link ou button) é o que faz esta asserção continuar
    // válida se a ação mudar de elemento HTML de novo — foi exatamente
    // essa mudança (botão → link, specs/_padrao-formularios.md) que
    // esvaziou a versão anterior desta checagem.
    const acaoDeEdicaoNaLinhaDoInep = linhaDoInep
      .getByRole("link", { name: /^(Editar|Inativar|Reativar|Excluir)/ })
      .or(linhaDoInep.getByRole("button", { name: /^(Editar|Inativar|Reativar|Excluir)/ }))
    await expect(acaoDeEdicaoNaLinhaDoInep).toHaveCount(0)

    const linhaPropria = page.getByRole("row", { name: /Reuniões com representação discente/ })
    await expect(linhaPropria).toBeVisible()
    await expect(linhaPropria).toContainText("Próprio")
    await expect(linhaPropria.getByRole("link", { name: "Editar GEST-01" })).toBeVisible()

    await logout(page)

    await login(page, {
      instituicao: "Instituto Vale Verde (IVV)",
      email: "renata.coimbra@ivv.edu.br",
      senha: "nde-vale-verde-26",
    })
    await expect(page).toHaveURL("/app")

    await page.goto("/app/indicadores")
    await page.getByRole("button", { name: "Pesquisar" }).click()

    await expect(page.getByRole("row", { name: /Núcleo Docente Estruturante/ })).toBeVisible()
    // GEST-01 é homônimo entre FSA e IVV (IN-03) — se a exceção do
    // catálogo comum vazasse o indicador próprio de outra instituição,
    // esta linha apareceria duas vezes em vez de uma.
    await expect(page.getByRole("row", { name: /GEST-01/ })).toHaveCount(1)

    await logout(page)
  })

  test("Administrador vê a contagem total sem identificar instituições, e não alcança /app/metas", async ({
    page,
  }) => {
    const senhaOriginal = "plataforma-2026"
    const senhaTemporaria = "plataforma-2026-e2e-catalogo"
    let trocouSenha = false

    // Rafael é o admin de bootstrap (SEED_ADMIN_EMAIL, design.md §9,
    // 3.13) — só nasce com senha provisória na primeira vez que o seed
    // roda; depois que alguém a troca, o self-service (`POST /auth/
    // senha`) grava `senha_provisoria=false` para sempre — não existe
    // caminho de volta sem outro administrador (E-12 impede
    // autorredefinição), e criar um só para isso, sem um terceiro para
    // excluí-lo depois, trocaria "resíduo de senha provisória" por
    // "resíduo de administrador" (E-08 impede autoexclusão — o ciclo
    // não fecha com só dois administradores). Por isso o teste aceita
    // os dois estados de entrada e só restaura o valor da senha, nunca
    // a flag — a asserção de "força a troca no primeiro acesso" já tem
    // dono em auth/primeiro-acesso.spec.ts, com um usuário comum.
    try {
      await login(page, { email: "rafael.toledo@basis-avalia.local", senha: senhaOriginal })

      // `toHaveURL` com timeout curto — não `locator.isVisible()`, que não
      // reespera a navegação que `login()` acabou de disparar e pode
      // checar a página antes do redirecionamento acontecer.
      let precisaDefinirSenha = true
      try {
        await expect(page).toHaveURL("/app/alterar-senha", { timeout: 3000 })
      } catch {
        precisaDefinirSenha = false
      }
      if (precisaDefinirSenha) {
        await expect(page.getByRole("heading", { name: "Defina sua senha" })).toBeVisible()
        await page.getByLabel("Senha atual", { exact: true }).fill(senhaOriginal)
        await page.getByLabel("Nova senha", { exact: true }).fill(senhaTemporaria)
        await page.getByLabel("Confirme a nova senha", { exact: true }).fill(senhaTemporaria)
        await page.getByRole("button", { name: "Salvar" }).click()
        trocouSenha = true
      }
      await expect(page).toHaveURL("/app")

      await page.goto("/app/indicadores-inep")
      await page.getByRole("button", { name: "Pesquisar" }).click()

      const linha = page.getByRole("row", { name: /Núcleo Docente Estruturante/ })
      await expect(linha).toBeVisible()

      const corpoDaPagina = await page.locator("body").innerText()
      for (const identificadorDeInstituicao of [
        "Faculdade Serra Azul",
        "Instituto Vale Verde",
        "(FSA)",
        "(IVV)",
      ]) {
        expect(corpoDaPagina).not.toContain(identificadorDeInstituicao)
      }

      await page.goto("/app/metas")
      await expect(page).toHaveURL("/app")
      await expect(page.getByText("Você não tem permissão para acessar esta área.")).toBeVisible()
    } finally {
      if (trocouSenha) {
        const restaurada = await page.request.post(`${API_URL}/auth/senha`, {
          data: { senha_atual: senhaTemporaria, senha_nova: senhaOriginal },
        })
        expect(restaurada.ok(), "restauração: senha de Rafael de volta ao valor do seed").toBeTruthy()
      }
      await page.request.post(`${API_URL}/auth/logout`)
    }
  })
})
