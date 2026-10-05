# Testes pendentes — autenticacao-usuarios

Este arquivo existe por **decisão de custo do dono do produto**, não por
esquecimento: a partir dos Grupos 6–11, a suíte de testes foi reduzida ao
mínimo necessário (isolamento entre instituições, invariantes com corrida,
matriz de autorização, concorrência otimista, unicidade, smoke de status
code). Caminho feliz de CRUD, paginação/ordenação/filtro, validação de
formato de campo, estados de tela/microcópia/dirty state e testes de
componente de frontend foram **adiados deliberadamente** para a fase de
Release, quando a suíte E2E Playwright for escrita — os requisitos ainda
podem mudar antes disso, e testar agora seria custo jogado fora.

Formato de cada linha: **código do cenário · o que verificar · camada · por
que foi adiado.**

> **Modelo de múltiplos perfis concluído em 2026-09-28.** O aviso que existia
> aqui sobre os Grupos 7 a 10 estarem implementados contra o modelo antigo
> (`perfil` único) foi resolvido: o Grupo 13 trocou o modelo para
> `ConjuntoDePerfis` em todo o backend, e os Grupos 7 a 10 foram atualizados
> na mesma rodada — controle de perfis do modal virou grupo de caixas de
> seleção (T-093), coluna do grid virou "Perfis" com resumo "+N" (T-094),
> breadcrumb existe em toda página autenticada (T-068), sincronização de
> perfis na sessão sem polling (T-092). As linhas abaixo para os Grupos 7 a
> 10 continuam válidas como mapa de cenários pendentes — agora contra o
> modelo atual, não mais um modelo que ia mudar.

## Login — correção fora do ciclo normal

- **Combo de instituição, clique após abrir o popup** · confirmar em
  navegador real que clicar em "Administração do sistema" ou em uma
  instituição da lista seleciona o item, fecha o popup e preenche o campo,
  sem o aviso de hidratação no console ("A tree hydrated but some
  attributes...") · componente frontend (`combobox-entidade.tsx`,
  `instituicao-login-combo.tsx`) · corrigido em 2026-09-28 (o item fixo
  "Administração do sistema" era renderizado como `ComboboxItem` fora da
  coleção rastreada pelo `Combobox.Root`, sem participar do registro de
  índice/id da lista — causa confirmada lendo o código-fonte de
  `@base-ui/react/combobox` e comparando com os exemplos oficiais de item
  fixo/"creatable"; corrigido movendo o item fixo para dentro de `items`,
  como um grupo próprio, com `filter` customizado para nunca ser removido
  pela busca). **Coberto** desde 2026-09-28 por `e2e/auth/login-logout.spec.ts`
  (clique real em "Administração do sistema" e em uma instituição da lista,
  em navegador Chromium de verdade, rodando contra `DATABASE_URL_TEST`) —
  o smoke manual do dono do produto deixou de ser necessário para este
  cenário.
- **L-09** (raiz do sistema é a tela de login) · **coberto implicitamente**
  — todo teste de `e2e/` começa com `page.goto("/")` e interage com o
  formulário de login; se a raiz não mostrasse login, a suíte inteira
  falharia na primeira linha. Nenhuma asserção dedicada, mas a suíte não
  passa sem esse comportamento estar correto.
- **L-10** (visitante sem sessão não acessa a aplicação, toast "Faça login
  para continuar.") · componente frontend · **sem teste dedicado** — o
  guarda de rota existe no código, mas nenhum teste tenta abrir
  `/app/usuarios` sem sessão e confirma o redirecionamento com esse toast
  específico. Registrado como pendência real, não coberta.
- **S-04** (confirmação de senha divergente, validação só no cliente) ·
  componente frontend · validação de formato de campo, adiada.

## Logout — bug real encontrado e corrigido em 2026-09-28

- **`useLogout` não navegava nem limpava estado ao falhar a chamada à API**
  · `features/auth/hooks/use-logout.ts` · a chamada `POST /auth/logout` não
  tinha `catch`: se a rede caísse ou a API respondesse 5xx, a exceção
  propagava, a navegação para `/` nunca era alcançada e a tela ficava presa
  em estado autenticado com o cookie já potencialmente inválido no
  servidor. Corrigido com `try/catch/finally`: o `finally` sempre limpa o
  motivo de encerramento guardado em `sessionStorage` e redireciona para
  `/` via `window.location.replace("/")` — `.replace` (não `.push`/`.href`)
  para que o histórico do navegador não tenha uma entrada da tela
  autenticada entre a sessão encerrada e o login, o que também é o que o
  teste de "botão voltar" abaixo comprova. **Coberto** por
  `e2e/auth/login-logout.spec.ts` (inclui um teste específico: sair e
  clicar em voltar não deve mostrar tela autenticada) e exercitado de
  passagem em todos os outros arquivos de `e2e/` que fazem login seguido de
  logout.

## Sessão após troca de senha — bug real encontrado pelo E2E em 2026-09-28

- **`useEu()` ficava com dado desatualizado depois de `POST /auth/senha`**
  · `features/auth/hooks/use-eu.ts` · dois consumidores independentes do
  hook (a tela de troca de senha e o guard de rota) cada um mantinha seu
  próprio estado local carregado no mount; depois de trocar a senha, o
  guard de rota ainda enxergava `senha_provisoria: true` de uma leitura
  anterior e mandava o usuário de volta para a tela de troca, que por sua
  vez já enxergava a sessão nova e tentava seguir adiante — um vaivém
  visível na tela. Corrigido trocando `useEu()` para uma store compartilhada
  em nível de módulo (assinantes ouvindo um único estado, sem requisição
  duplicada) e expondo `definirSessaoAtual()` para `alterar-senha-form.tsx`
  publicar a sessão nova imediatamente após o sucesso da troca, antes de
  qualquer redirecionamento. Este bug não foi reportado pelo dono nem
  existia como cenário pendente conhecido — apareceu ao escrever
  `e2e/auth/primeiro-acesso.spec.ts`, exatamente o tipo de defeito que só
  aparece com navegador real. **Coberto** por
  `e2e/auth/primeiro-acesso.spec.ts` (login com senha provisória → tela de
  troca → salvar → chega em `/app` sem vaivém).
- **S-01** (primeiro acesso exige troca de senha) — não tinha teste, nem Go
  nem menção neste arquivo, até esta rodada de correção. **Coberto
  parcialmente** por `e2e/auth/primeiro-acesso.spec.ts`: confirma que a
  senha provisória leva a "Defina sua senha" e que a troca funciona.
  **Não verificado**: tentar navegar para outra tela durante a senha
  provisória devolve para lá (o guarda existe no código —
  `middleware_sessao.go` + o guarda de rota do frontend — mas nenhum teste
  automatizado tenta navegar e confirma o bloqueio). Registrado como
  pendência específica, não coberta.

## Grupo 6 — Instituições

- **I-03** · nome e sigla obrigatórios, mensagens exatas · handler ·
  validação de formato de campo, adiada.
- **I-06/I-07** · inativar/reativar devolve 200, `situacao` muda, nenhum
  usuário é tocado, sessões da instituição caem na próxima requisição ·
  handler + integração · caminho feliz + já coberto indiretamente pelo
  teste de middleware de sessão do Grupo 3 (INSTITUICAO_INATIVA).
- **I-08** · **correção de rótulo desta rodada**: a linha anterior aqui
  ("`pesquisadores_ativos` aparece corretamente por linha no grid") não é
  o enunciado atual de `I-08` em `spec.md` — é
  `TestInstituicaoRepository_I08_ContagemDePesquisadoresAtivosPorLinha`,
  já coberto, mas com a tag trocada (mesma família de achado de
  AS-05/AS-09 abaixo). O `I-08` **atual** ("instituição sem PI é destacada
  no grid e fica fora do combo") está **parcialmente** coberto: a metade
  "fica fora do combo público" tem teste de repositório
  (`TestInstituicaoPublicaQuery_L11_SoAtivaComPIAtivo`, também com a tag
  L-11 em vez de I-08). A metade "destacada no grid, com o texto de apoio"
  é **frontend, sem teste — pendência real, não coberta**.
- **I-10** · ordenação padrão `nome asc`, allowlist inclui `codigo_emec` ·
  handler · paginação/ordenação, adiada (allowlist em si é testada no
  handler apenas por smoke de 400 com `sort` inválido).
- **I-12** · a metade de backend (`409 CONFLITO_DE_VERSAO` ao salvar com
  `versao` desatualizada) está **coberta** por
  `TestSmoke_Instituicoes_CicloCompletoDeStatusCodes`. A metade "mensagem
  de conflito de versão na tela" (frontend/E2E) segue **adiada** — sem
  teste.
- Formulário de instituição (modal): dirty state, dois-cliques, foco inicial
  · componente frontend · adiado para o Grupo 11/Release (Vitest fica só
  configurado, sem suíte ampla).
- **V-3** · o planner do Postgres de fato usa `idx_usuario_perfil_pi` (em vez
  de `Seq Scan`) na sonda do combo público de instituições · integração
  (`EXPLAIN ANALYZE`) · adiado porque a massa de um teste de integração
  (poucas linhas) torna `Seq Scan` a escolha correta e mais barata do
  planner — um teste que afirma uso de índice nesse volume falharia sempre
  no ambiente de teste e passaria em produção. `TestInstituicaoPublicaQuery_
  V3_IndiceParcialExiste` (`instituicao_publica_query_test.go`) verifica
  hoje, via `pg_indexes`, apenas que a migration criou o índice parcial
  correto — não que o planner o escolhe. Reavaliar quando houver volume
  representativo de instituições/usuários em ambiente de teste, ou quando a
  consulta cruzar o gatilho de latência que o arquiteto definiu para entrar
  em cache (acima de 10 ms) — o que vier primeiro.

## Grupo 7 — Pesquisadores Institucionais

- **AS-01** · criar o primeiro PI e a instituição passar a aparecer no
  combo público no mesmo teste (E2E completo do fluxo) · integração · a
  parte de isolamento entre instituições agora tem cobertura E2E real —
  `e2e/usuario/isolamento.spec.ts` (2026-09-28), não mais só o teste de
  repositório da suíte mínima. **Ainda pendente**: o fluxo de tela
  encadeado em si — criar instituição, o modal "Cadastrar o primeiro
  Pesquisador Institucional" abrir automaticamente, completar e a
  instituição aparecer no combo público — não foi automatizado (fora da
  lista de prioridade do dono para esta rodada de E2E); segue adiado.
- **E-10** · com dois PIs, rebaixar um é aceito · use case · caminho feliz
  do lado "sem violar a invariante"; a invariante em si (última) tem teste
  de corrida mantido.
- **G-17 (variante do grupo)** · `/app/instituicoes/[id]/pesquisadores`
  carrega a lista ao montar, sem botão "Pesquisar" (exceção de design.md
  §6.2 R17) · componente frontend · estado de tela, adiado.
- Modal encadeado "Cadastrar o primeiro Pesquisador Institucional de
  {nome}": abre automaticamente ao salvar uma instituição nova; campo
  Perfil aparece como texto fixo, não `Select`; fechar sem completar deixa
  o badge/botão "Sem Pesquisador Institucional" permanente na linha do
  grid de instituições · componente frontend/E2E · estado de tela e fluxo
  encadeado, adiado.
- Link do nome da instituição na tabela/card leva a
  `/app/instituicoes/[id]/pesquisadores` · componente frontend · navegação,
  adiado.

## Grupo 8 — Usuários da instituição (PI)

- **U-02, U-03, U-09** · e-mail duplicado na mesma instituição, e-mail de
  excluído volta a ficar livre, mesmo e-mail em duas instituições ·
  repositório · **coberto** por `TestUsuarioRepository_U02_
  EmailDuplicadoNaMesmaInstituicao`, `_U03_EmailDeExcluidoVoltaAFicarLivre`
  e `_U09_MesmoEmailEmDuasInstituicoes` (rodada de 2026-09-29 — fronteira
  de integridade que a §15.1 exige e que só existia como índice, sem teste
  que provasse o que ele promete).
- **U-04 a U-08** (exceto U-10, coberto por `conjunto_de_perfis_test.go`,
  e U-11, coberto por
  `TestCriarUsuario_U11_InstituicaoVemDaSessaoNuncaDoPayload`) ·
  validação de campo, duplo clique · handler/frontend · adiado.
- **G-19** (coluna "Perfis" mostra todos, com truncamento "+N" em ordem
  alfabética quando não couber) · componente frontend · `e2e/usuario/
  crud.spec.ts` confirma que os dois perfis marcados aparecem na linha, mas
  o usuário de teste tem só 2 perfis — não aciona o truncamento. Sem teste
  do caso de 3+ perfis. Adiado.
- **G-01, G-05, G-06, G-11 a G-16, G-18** · estados de tela (abre sem
  consulta, alternância de ordenação por clique, estado restaurado do
  localStorage, loading, falha na pesquisa, primeiro acesso pendente na
  linha) · componente frontend · adiado. **Correção de 2026-09-28**: uma
  rodada anterior deste arquivo dizia "G-01 a G-18 (exceto G-08/G-10)" —
  incompleto. G-02 (busca por nome/e-mail), G-03 (filtro por perfil), G-04
  (ordenação padrão), G-07 (coluna não ordenável), G-09 (paginação) e G-17
  (excluídos não aparecem) também têm teste de repositório/query dedicado,
  além de G-08/G-10 — são concerns de **consulta** (backend), diferentes
  dos itens desta lista, que são de **interação de tela** (frontend).
  Ressalva: `e2e/usuario/crud.spec.ts` toca de passagem o caso mais básico
  do grid (linha nova aparece com os perfis corretos após criar, some após
  excluir), mas não os itens acima.
- **E-06, E-07** · confirmação de exclusão nomeando quem será excluído (e
  "Cancelar" não altera nada), loading isolado por linha no grid · handler/
  frontend · **correção de 2026-09-28**: uma rodada anterior deste arquivo
  reivindicou os dois como cobertos por `e2e/usuario/crud.spec.ts` — não é
  verdade. Esse teste confirma que existe uma barreira de confirmação antes
  de excluir (não é possível excluir com um clique) e que a exclusão de
  fato acontece depois de confirmar, mas **não verifica** (a) o texto do
  diálogo nomeando a pessoa, (b) que "Cancelar" preserva o registro, nem
  (c) que só o botão daquela linha entra em estado de carregamento
  enquanto as outras permanecem normais. Revertido para **adiado**.
- **E-02, E-13** · auditoria de campos alterados, dirty state do grupo de
  checkboxes · handler/frontend · adiado.
- Modal de usuário (`usuario-form-modal.tsx`): "Novo" com nome, e-mail, grupo
  de checkboxes de perfis, senha, sem campo de instituição; "Editar" sem
  senha; grupo de checkboxes desabilitado com `FieldDescription` explicativa
  ao editar o próprio registro; mensagem de e-mail duplicado "nesta
  instituição" exibida inline no campo, não em toast · componente frontend
  · estado de tela, adiado.
- **E-16 (interação)** · desmarcar o último perfil marcado remarca "Aluno" na
  hora, com o texto explicativo, sem round-trip ao servidor · componente
  frontend (`usuario-form-modal.tsx`) · comportamento verificado por leitura
  de código (a função `handlePerfisChange` trata o caso), não por teste de
  componente — sem Vitest/Playwright de componente configurado nesta fase.
- **E-13 (dirty state do grupo de checkboxes)** · marcar/desmarcar qualquer
  perfil conta como alteração para o aviso de "alterações não salvas" ·
  componente frontend · mesma razão acima.
- Navegação por teclado do grupo de checkboxes (`Tab` percorre, `Espaço`
  marca, rótulo do grupo é lido pelo leitor de tela) · componente
  frontend/acessibilidade · adiado para a fase de Release (Playwright +
  axe-core).
- `redefinir-senha-modal.tsx`: duas senhas devem coincidir, aviso de senha
  provisória, oculta na própria linha (E-12) · componente frontend ·
  validação de formulário, adiada.

## Grupo 9 — Administradores do Sistema

- **AS-08** · quem não é administrador não alcança `/instituicoes` ·
  handler/frontend · a metade de backend (API responde 403
  `PERMISSAO_NEGADA`) **já coberta** por
  `TestSmoke_Instituicoes_PINaoAlcanca_403` (sem o código na tag do nome —
  achado desta rodada, não corrigido para não mexer em teste alheio sem
  necessidade). A metade de frontend (toast de permissão negada ao digitar
  a URL) segue adiada.
- **AS-04** · o menu do administrador mostra só o grupo "Sistema" ·
  componente frontend · estado de tela, adiado.
- **AS-05** · vínculo institucional nulo, cabeçalho/rodapé mostrando
  "Administração do sistema" sem sigla, ausência em grids de instituição ·
  handler/frontend · a metade "ausência em grids de instituição" seria
  coberta pelo mesmo isolamento que T-01 já testa estruturalmente
  (query nunca mistura os dois universos, design.md §4.4), mas não há teste
  que crie um administrador e confirme sua ausência explícita num grid de
  usuários de instituição. A metade de cabeçalho/rodapé é frontend, adiada.
  **Sem teste dedicado — registrado como pendência real, não coberta.**
  **Nota de rotulagem** (achado de rodada anterior, preservado aqui):
  `TestSmoke_Administradores_AS05_ListaTrazSoAdministradores` tem a tag
  `AS05` no nome, mas o que a função exercita de fato é o enunciado de
  `AS-09` (grid de administradores não lista usuário de instituição) — já
  coberto sob esse rótulo trocado. Não renomeada (teste de outra rodada,
  comportamento correto, só o nome está errado) — quem for mexer nesse
  arquivo depois não deve se confundir: a pendência real de `AS-05` acima
  continua sem teste dedicado.
- Tela `/app/administradores`: reaproveita `usuario-table`/`usuario-filtro`/
  `usuario-form-modal` de `features/usuario` com `perfilFixo=
  "administrador_sistema"` e sem filtro de perfil (`mostrarFiltroPerfil=
  false`); redefinir senha também reaproveitado (rota R23-R28 já cobre
  `POST /administradores/{id}/senha`) · componente frontend · caminho feliz
  e reaproveitamento, adiado.
- Exclusão do último administrador (`409 ULTIMO_ADMINISTRADOR_SISTEMA`)
  mostrando o toast de erro devolvido pela API · componente frontend ·
  a invariante em si já tem teste de corrida (T-063); só a exibição do
  toast está adiada.

## Grupo 10 — AppShell completo

> **Estado em 2026-09-28: concluído.** `nav-config.ts`/`app-sidebar.tsx`/
> `sidebar-group.tsx`/`sidebar-nav.tsx`, `use-guarda-de-permissao.ts`
> (T-066), toggle de colapso, `Drawer` mobile, atalhos de teclado, e agora
> também `trilha.tsx` (T-068, breadcrumb em toda página autenticada) e
> `use-sincronizacao-de-perfis.ts` (T-092) — todos implementados, TypeScript
> e ESLint limpos, `next build` compila as 9 rotas sem erro.
- **`SH-01`** (regiões do shell) · **`SH-03`** (breadcrumb, incluindo elipse
  em 360px) · **`SH-04`** (grupo do menu expansível, item ativo destacado) ·
  **`SH-05`** (menu recolhido, ícones com dica) · **`SH-06`** (barra de
  progresso na navegação) · **`SH-07`** (barra de progresso em chamada ao
  servidor) · **`SH-08`** (notificações: cor/ícone por tipo, empilhamento,
  fechar antes do tempo, erro mais longa que sucesso) · **`SH-09`**
  (conteúdo alinhado à esquerda em 1440px) · **`SH-10`** (responsividade em
  360/768/1440px) · **`SH-11`** (navegação por teclado: ordem de tabulação,
  foco visível, Espaço no grupo de perfis, Esc com proteção de dados não
  salvos) · **`SH-12`** (rodapé identifica a instituição da sessão) ·
  **`SH-13`** (sigla da instituição no cabeçalho) — menu hierárquico por
  permissão, breadcrumb, atalhos de teclado, responsividade nas três
  larguras, `aria-live` dos grids · frontend/E2E · adiado para a fase de
  Release — é o grupo mais dependente de interação visual real, que só faz
  sentido testar com Playwright.
- **SE-03/SE-09 (sincronização de perfis ao vivo)** · perfil retirado durante
  a sessão faz o item de menu sumir na troca de rota seguinte, sem sair e
  entrar; perfil acrescentado faz o item aparecer; nenhum recarregamento de
  página · componente frontend/E2E · a lógica de comparação e o toast estão
  implementados (`use-sincronizacao-de-perfis.ts`), mas o comportamento
  ao vivo (dois perfis mudando no banco enquanto a sessão está aberta no
  navegador) só é verificável com um teste E2E real ou smoke manual —
  adiado para a fase de Release.
- Revisão visual do breadcrumb nas três larguras (elipse abaixo de 360px) ·
  componente frontend · adiado para a fase de Release.

## Bug real encontrado ao rodar a suíte de integração Go junto com o E2E

- **`basisavalia_test` e o seed do E2E compartilhando o mesmo banco
  quebrava `TestExcluirUsuario_AS06_ConcorrenciaDoUltimoAdministrador`**
  · descoberto ao rodar a suíte de integração Go (`docker compose ... run
  test`) depois de configurar a infraestrutura de E2E desta rodada —
  8 em 8 execuções isoladas falhavam de forma **determinística** (não
  era flake): "esperava exatamente 1 sucesso e 1 conflito, obtido 2
  sucesso(s) e 0 conflito(s)". Causa: `migrate-e2e`/`seed-e2e`/
  `backend-e2e` apontavam para o mesmo `DATABASE_URL_TEST` que o serviço
  `test` (Go) usa — o seed do E2E cria um Administrador do Sistema
  persistente (não limpo entre execuções, por design: é dado de
  referência estável para os testes de navegador). O teste AS-06 cria
  exatamente 2 administradores e espera que remover um dos dois dispare
  a invariante "não pode ficar sem nenhum" ao remover o segundo — mas com
  o administrador da plataforma do E2E também presente na mesma tabela, a
  contagem nunca chega a zero, então as duas remoções concorrentes
  sempre "sucedem". **Corrigido** dando ao E2E um banco totalmente
  próprio: `DATABASE_URL_E2E` (`basisavalia_e2e`), configurado em
  `docker/postgres/init-test-db.sh`, `.env`/`.env.example` e
  `docker-compose.dev.yml`. Confirmado depois da correção: suíte Go
  completa **passa** (`internal/usecase/command/usuario` `ok`, AS-06
  incluído) e a suíte E2E continua passando contra o próprio banco.

## Intermitência de sessão sob carga — bug real encontrado e corrigido em 2026-09-29

- **`relogio.Relogio.Agora()` não era monotônico, e `testhelpers.CriarUsuario`
  não passava pelo `Relogio` da aplicação** · `internal/adapter/relogio/
  relogio.go`, `internal/testhelpers/fixtures.go` · descoberto investigando
  `TestSmoke_R5_CookieNovoFuncionaECookieAntigoExpira` e outros quatro testes
  (`T02`, `T05`, `AS03`, `AS10`, `G07`) falhando de forma intermitente **só**
  sob a suíte completa (`go test ./...`), nunca isolados — todos comparam
  `emt` do cookie contra `sessoes_validas_a_partir_de` gravado por outra
  chamada de relógio. Causa raiz provada por três camadas, cada uma
  confirmada com evidência direta antes de corrigir (`systematic-debugging`,
  nunca "rodei de novo e passou"):
  1. `time.Now()` cru na fixture divergia por arredondamento de 1µs do `emt`
     truncado do login (Postgres arredonda o resto de nanossegundo ao
     gravar; `UnixMicro()` trunca) — corrigido truncando a fixture também.
  2. **O relógio de parede do container (WSL2/Docker Desktop) retrocede de
     verdade sob carga de CPU** — provado com um microbenchmark Go isolado
     (sem HTTP, sem banco, sem nenhum arquivo de teste tocado): amostragem
     pura de `time.Now()` sob carga de CPU equivalente à do argon2 capturou
     **1 retrocesso de ~547ms em 83 milhões de amostras em 20s**. Duas
     chamadas de `Agora()` em pontos diferentes do fluxo (login, depois
     troca de senha) podiam produzir um `emt` maior que o valor gravado
     depois. Corrigido: `Relogio.Agora()` agora garante monotonicidade
     (nunca retrocede em relação à própria última leitura).
  3. Mesmo com `Agora()` monotônico, cada chamador fazia `relogio.Novo()`
     — uma instância nova, sem memória do retrocesso de nenhuma outra. A
     fixture (chamada fora de qualquer `Relogio`) e o app de cada teste HTTP
     (`montarAppDeTeste`, uma instância própria por teste) não se protegiam
     um do outro. Corrigido: `Novo()` passa a devolver **sempre a mesma
     instância dentro do processo** (memoizada com `sync.Once`), e a
     fixture passou a usar essa mesma instância em vez de `time.Now()` cru.
  **Prova de correção**: suíte completa (`go test ./... -count=1`) rodada
  **5 vezes consecutivas**, sem nenhum arquivo editado durante as execuções
  e sem nenhum outro processo concorrente contra `basisavalia_test`, **as
  5 totalmente verdes**. Reprodução adicional de 200 iterações dos 6 testes
  historicamente afetados, em pacote isolado: 0 falhas antes do harness
  atingir o timeout padrão de 10 minutos do `go test` (não uma falha real —
  `panic: test timed out after 10m0s`, esperado com esse volume de
  requisições HTTP+argon2 sequenciais).
  **Recomendação registrada para o `arquiteto`**: o serviço `test` do
  `docker-compose.dev.yml` monta `./backend:/app` **ao vivo** — o mesmo
  diretório editado durante o desenvolvimento. Editar um `.go` durante a
  fase de compilação de `go test ./...` pode fazer o compilador ler fonte
  parcialmente escrita, produzindo erro/comportamento transitório em
  pacotes aparentemente aleatórios — um segundo mecanismo de intermitência,
  **independente** do de relógio acima (este projeto já tem o precedente de
  `backend-e2e` migrado para build a partir de cópia estática do código,
  por um problema relacionado). Não fazia parte desta correção porque a
  prova das 5 execuções verdes já isola esse fator (nenhuma edição durante
  as execuções), mas o `test` do backend continua vulnerável a esse
  problema para qualquer sessão futura de desenvolvimento em paralelo com
  a suíte — considerar o mesmo tratamento dado ao `backend-e2e`.

  **Confirmado de forma independente pelo coordenador em 2026-09-29**: cinco
  execuções da suíte completa **sem ninguém editando** — todas verdes; execuções
  anteriores **com edição concorrente** — falha em praticamente toda execução,
  testes diferentes a cada vez, sem padrão. Isso descarta de vez qualquer
  explicação de estado sujo ou ordem de teste para aquelas falhas: o
  compilador lia fonte parcialmente escrita, não havia defeito nem no
  teste nem no código de produção. **Regra a partir de agora**: nunca rodar
  `docker compose ... run test` (suíte completa) enquanto qualquer arquivo em
  `backend/` está sendo salvo — os dois não convivem com o código montado ao
  vivo. A correção estrutural (build a partir de cópia estática, como
  `backend-e2e` já faz) é decisão do `arquiteto`, registrada aqui e **não
  implementada** por conta própria.

## `e2e/auth/primeiro-acesso.spec.ts` não era re-executável — bug de teste encontrado e corrigido em 2026-09-29

- **`senha_provisoria` de `joao.ribeiro@ies.edu.br` (FSA) ficava `false`
  permanentemente depois da primeira execução bem-sucedida** ·
  `e2e/auth/primeiro-acesso.spec.ts` · investigado a pedido do coordenador
  para descartar defeito de backend antes de mexer no teste — **confirmado
  que não é defeito de backend**: `AlterarSenhaPropria` sempre grava
  `senha_provisoria=false` (design correto, sem mecanismo de "desfazer");
  o teste restaura o **valor** da senha ao final (`POST /auth/senha` de
  volta para "senha-fsa-2026"), mas nunca restaura a **flag** de
  provisória, que só volta a `true` via `RedefinirSenha` (rota de
  administração, não usada pelo teste). Como `basisavalia_e2e` persiste
  dados entre execuções por desenho, e `criarPessoaSeNaoExistir` do seed é
  idempotente (nunca sobrescreve linha existente — correto para não
  mascarar dado real), a segunda execução em diante do teste sempre falhava
  determinística e permanentemente, não de forma intermitente. Confirmado
  lendo `senha_provisoria` direto no banco (`f` nas duas linhas de
  `joao.ribeiro@ies.edu.br`) antes de qualquer alteração de código.
  **Corrigido** com `test.afterEach` em `primeiro-acesso.spec.ts`: loga
  como `maria.souza@fsa.edu.br` (PI da FSA, já no seed), localiza
  `joao.ribeiro@ies.edu.br` via `GET /usuarios?busca=`, e chama
  `POST /usuarios/{id}/senha` (a mesma rota de `RedefinirSenha`) para
  devolver a conta ao estado de primeiro acesso — roda **sempre**, mesmo
  se o teste falhar antes de chegar à restauração própria. **Prova**:
  rodado 3 vezes seguidas depois da correção (1ª ainda falhou, consumindo
  o estado herdado de antes da correção; 2ª e 3ª passaram), e a suíte
  `e2e/` inteira (10 testes, todos os arquivos) passou de ponta a ponta
  depois.

## Grupo 11 — Fechamento

> **Estado em 2026-09-28: concluído.** `/version` (T-071), CSP com nonce em
> `proxy.ts` (T-070), swaggo em todos os 29 handlers com `docs/swagger.yaml`
> gerado e servido em `/swagger/index.html` (T-072), e as quatro ferramentas
> de qualidade do frontend instaladas e **rodando de verdade**, não só
> presentes no disco (T-074) — ver `biome.json`, `knip.config.ts`,
> `.dependency-cruiser.js`, `vitest.config.ts`, `commitlint.config.js` +
> `.husky/commit-msg` na raiz do repositório.

- **CSP em produção sem erro no console** · manual/E2E · a política com
  nonce está implementada em `frontend/src/proxy.ts`, aplicada só em
  `APP_ENV=production` (relaxada em dev, conforme CLAUDE.md). **Verificado
  por requisição HTTP direta** contra uma build de produção real
  (`next build` + `node .next/standalone/server.js`, `APP_ENV=production`):
  todas as 8 rotas do App Router são `ƒ` (dinâmicas) — corrigido nesta
  rodada com `export const dynamic = "force-dynamic"` em `app/layout.tsx`,
  depois de descobrir que 7 das 8 rotas nasciam estáticas por padrão, o que
  quebraria a aplicação de nonce (a página estática não tem nonce para
  aplicar aos scripts do Next, gerado só por requisição). Os 6 cabeçalhos
  (`Content-Security-Policy` com nonce real, `X-Frame-Options`,
  `X-Content-Type-Options`, `Strict-Transport-Security`, `Referrer-Policy`,
  `Permissions-Policy`) chegam corretos na resposta. **Não verificado**:
  console do navegador real sem violação de CSP ao interagir com a
  aplicação — só a presença e a forma do cabeçalho foram confirmadas, não
  a ausência de bloqueio de algum script/estilo específico em uso. Smoke
  manual contra `docker compose -f docker-compose.yaml up --build` antes
  do primeiro deploy — adiado para a fase de Release.

### Débito de lint descoberto pelas ferramentas novas — decisão registrada

Duas ferramentas começaram a rodar nesta rodada (ESLint já rodava; Biome é
novo) e **ambas apontam a mesma causa raiz**: efeitos que buscam dado ou
mantêm estado local sem uma biblioteca de dados (`@tanstack/react-query`),
cujo gatilho de adoção já está registrado em `design.md` §11.5 ("o primeiro
candidato real agora é o badge de recusas de `metas-coordenacao`").

- **ESLint** — `react-hooks/set-state-in-effect`: **18 erros, 2 avisos**,
  em ~7 arquivos. Contagem estável antes e depois desta rodada (nenhum
  novo introduzido; um caso foi genuinely corrigido em `use-eu.ts`, sem
  mudar o total porque já estava suprimido por comentário, não contado).
- **Biome** (primeira execução do projeto) — **49 problemas restantes**
  depois de aplicar os fixes seguros (`--write`, sem `--unsafe`) e corrigir
  manualmente 4 casos comprovadamente seguros (`use-eu.ts`, `field.tsx`
  `==`→`===`, `carregamento.ts` retorno de `forEach`, parser CSS do
  Tailwind v4 em `biome.json`):
  - `lint/correctness/useExhaustiveDependencies` em 9 arquivos
    (`app-shell.tsx`, `use-local-storage.ts`,
    `use-sincronizacao-de-perfis.ts`, os 4 `page.tsx` de listagem, os 2
    `*-form-modal.tsx`) — **verificado caso a caso, não presumido**: pelo
    menos dois (`app-shell.tsx` fechar o menu mobile ao trocar de rota;
    `use-sincronizacao-de-perfis.ts` comparar perfis ao trocar de rota)
    têm comentário `eslint-disable` **pré-existente e intencional** — o
    efeito depende de disparar exatamente quando `pathname` muda, e a
    correção sugerida pela ferramenta (adicionar a função ao array e
    remover `pathname`) **quebraria esse comportamento**. Aplicar
    `--unsafe` cegamente teria introduzido regressão, não corrigido débito.
  - `lint/a11y/*` em 4 componentes de `components/ui/` (`label.tsx`,
    `field.tsx`, `input-group.tsx`, `breadcrumb.tsx`) — primitivos que
    recebem `...props` do chamador; pelo menos o caso de `label.tsx`
    (`noLabelWithoutControl`) é provavelmente falso positivo, já que
    `htmlFor` é repassado pelos consumidores e a análise estática do Biome
    não enxerga isso. Precisa de revisão dedicada de acessibilidade, não
    de supressão às pressas.

**Decisão:** registrar como débito consolidado sob o mesmo gatilho de
`react-query` já aprovado pelo arquiteto — não corrigir agora. Corrigir
caso a caso agora exigiria reescrever o padrão de busca de dado em ~9
arquivos sem a biblioteca que resolveria o problema de raiz, ou suprimir
achados de acessibilidade sem a revisão que eles merecem. **O número não
sumiu por trás da troca de ferramenta**: ESLint continua em 18/2 (estável),
Biome estreou com 49, e os dois apontam para o mesmo gatilho de resolução.
Quando `react-query` entrar (gatilho já definido), os 9 casos de
`useExhaustiveDependencies` desaparecem por construção (o hook de dados
gerencia suas próprias dependências); os 4 casos de `lint/a11y/*` exigem
revisão própria, independente da adoção de `react-query` — registrar como
item separado quando o `ux-designer`/`security-reviewer` fizerem a
passagem de acessibilidade da fase de Release.
