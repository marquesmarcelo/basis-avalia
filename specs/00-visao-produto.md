# Visão Geral do Produto — basis-avalia

> Contexto compartilhado para todas as specs de funcionalidade — elas não
> precisam repetir o que está aqui.

**Data desta versão:** 28/09/2026
**Status:** aprovada pelo dono do produto

---

## O que este produto faz (1 parágrafo)

O basis-avalia apoia instituições de ensino no **acompanhamento das metas da
coordenação**, tomando como base os instrumentos de avaliação do INEP. A partir
de dois catálogos — os **indicadores**, derivados do instrumento, e as **metas**
da instituição, cada uma apontando para um ou mais indicadores — o Pesquisador
Institucional monta o **plano de ação de um curso em um período**, definindo
**quanto** daquele curso se exige de cada meta. Enquanto o plano está vigente, o
coordenador responde **enviando o arquivo comprobatório** de cada entrega, e o
Pesquisador Institucional avalia, aceitando ou recusando. O resultado é um
**relatório de desempenho por curso**, que mostra à instituição o que já foi
cumprido e o que falta, com trilha de quem entregou o quê e de quem aceitou ou
recusou. **Uma única instalação atende várias instituições**, com isolamento
completo de dados entre elas.

---

## O encadeamento das entidades

O modelo inteiro em uma figura, porque é ele que explica por que as features são
cinco e nesta ordem:

```
Indicador — escopo PLATAFORMA (catálogo do INEP)   ┐
Indicador — escopo INSTITUIÇÃO (próprio da IES)    ┘ N:N, pelo menos um
        ↑
      Meta (catálogo da instituição)
        ↑
Curso + Período ──> Plano de Ação Curso/Coordenador ──> Item do plano (meta + quantidade)
                            │                          │
                            └─> Documento DOCX          └─> Entrega ─> Anexos
```

Três leituras que precisam estar claras antes de qualquer spec:

- **O catálogo não sabe de curso nem de quantidade.** Uma meta é um enunciado
  reutilizável. É o **plano** que a transforma em cobrança concreta, ao atribuir
  curso, período e quantidade.
- **A quantidade é do item do plano, nunca da meta.** É assim que o mesmo
  enunciado — "registrar reuniões de NDE em ata" — pode exigir 4 de um curso e 6
  de outro.
- **Uma meta aponta para um ou mais indicadores, e uma entrega atende todos eles
  de uma vez.** Sem isso, a instituição duplicaria a meta para contar nos dois
  lugares e **passaria a exigir o dobro de entregas do coordenador** — o caso
  concreto é a ata de NDE, que atende o indicador de funcionamento do NDE e o de
  atuação do coordenador.

---

## Multi-institucionalidade (escopo da v1, não roadmap)

Uma instalação atende **várias instituições de ensino**. Isso atravessa todo o
modelo de dados e toda a autorização — não é camada adicionada depois.

- **Isolamento por coluna, banco único.** Toda tabela de entidade que pertence a
  uma instituição carrega o identificador dela.
- **O filtro por instituição é obrigatório em toda consulta e centralizado no
  adapter de persistência**, junto com o filtro de `excluido_em IS NULL`. Nunca
  repetido à mão em cada query — pelo mesmo motivo que o CLAUDE.md dá para a
  deleção lógica: a query onde alguém esquecer é a que vaza dado de outra
  instituição.
- **Isolamento é regra de autorização, não filtro de tela.** Avaliado no use
  case. **Recurso de outra instituição responde 404, nunca 403** — um 403
  confirmaria que o registro existe em outra instituição.
- **E-mail é único por instituição, não globalmente.** A mesma pessoa pode ter
  conta em duas instituições com o mesmo e-mail. Cada conta é independente.
- **A instituição é escolhida na tela de login**, em um combo carregado de rota
  pública. Só entram no combo instituições **ativas e com Pesquisador
  Institucional ativo**.
- **A instituição é fixada no login** e vale até o logout.

### A única exceção: o catálogo comum de indicadores do INEP

Os **indicadores de escopo plataforma são um catálogo único da instalação**,
mantido pelo Administrador do Sistema (feature `indicadores`). O instrumento é o
mesmo para todas as IES, e cadastrá-lo uma vez evita que cada instituição digite
os mesmos indicadores com números e nomes divergentes — o que tornaria
impossível comparar ou padronizar qualquer coisa depois. Cada instituição
continua criando **indicadores próprios**.

**Por que isso não fere o isolamento** — o argumento importa mais que a
conclusão, porque é ele que permite ao `arquiteto` julgar qualquer exceção
futura pelo mesmo critério:

- O filtro deixa de ser "a instituição da sessão" e passa a ser **"a instituição
  da sessão **ou** escopo de plataforma"**.
- **Linha de escopo plataforma não tem instituição.** Ela não pertence a
  ninguém, e por isso não pode ser de outra.
- Os dois conjuntos são **disjuntos**, e a união deles é **incapaz por
  construção** de retornar dado de outra instituição. Não é uma exceção vigiada
  por disciplina de quem escreve a query — é uma exceção que não tem como
  vazar, o que é uma propriedade diferente e muito mais forte.

**O que não atravessa, e é onde a exceção seria perigosa se fosse descuidada:**
nada sobre o **uso** do indicador cruza a fronteira. O Administrador do Sistema
vê apenas a **contagem total de uso** de um indicador do INEP, **nunca quais
instituições o usam**. Saber que a IES X adotou o indicador Y já seria
informação institucional dela, e não é o que ele precisa para manter um catálogo.

**A Meta não é compartilhada.** Ela é catálogo **da instituição**, mesmo quando
aponta para indicadores do INEP. A exceção alcança o indicador e para aí.

---

## Perfis: um usuário pode ter vários

**O modelo é de conjunto, não de valor único.** Uma pessoa possui um **conjunto
de perfis**, e as permissões dela são a **união** das permissões de cada um.

- **Qualquer combinação é permitida** entre **Aluno, Professor, Coordenador de
  Curso e Pesquisador Institucional.** Um professor que também responde
  institucionalmente pela IES é o caso comum em instituição pequena; o aluno de
  pós-graduação que também leciona é o caso que motivou o acúmulo. **Inclusive
  Pesquisador Institucional e Coordenador de Curso na mesma pessoa** — ver a
  decisão de segregação de funções abaixo.
- **Todo usuário nasce Aluno — como padrão quando nada é marcado, não como
  acréscimo.** Quem é cadastrado com "Professor" marcado fica apenas Professor.
- **Ninguém fica sem perfil.** Desmarcar o último recai para Aluno.
- **Administrador do Sistema é a única exceção: não acumula com nada.** O motivo
  é estrutural — ele é definido por **não pertencer a instituição alguma**,
  enquanto os outros quatro só existem **dentro de uma**. Um usuário com os dois
  precisaria ter e não ter instituição ao mesmo tempo, e todo o isolamento se
  apoia na instituição do ator. Quem administra a plataforma e também leciona
  tem **duas contas, com e-mails distintos**.
- **Coordenador de Curso deriva do vínculo com curso**, a partir da entrega de
  `cursos`: vincular alguém a um curso acrescenta o perfil; retirá-lo do último
  curso o remove, sem afetar os demais perfis.

| Rótulo na interface | Valor técnico | Pertence a instituição? | Acumula? |
|---|---|---|---|
| Administrador do Sistema | `administrador_sistema` | **não** | **não** |
| Pesquisador Institucional | `pesquisador_institucional` | sim | sim |
| Coordenador de Curso | `coordenador_curso` | sim | sim (por vínculo) |
| Professor | `professor` | sim | sim |
| Aluno | `aluno` | sim | sim |

**Consequência que atravessa todas as specs:** consultas e regras perguntam
**"quem possui o perfil X"**, nunca "quem é X". As invariantes de "ao menos um
Pesquisador Institucional ativo por instituição" e "ao menos um Administrador do
Sistema ativo" contam **detentores do perfil**.

### Segregação de funções: a coincidência é marcada, não proibida

Foi proposto impedir que a mesma pessoa fosse **Pesquisador Institucional e
Coordenador de Curso**, para que ninguém avaliasse a própria entrega. **O dono
do produto recusou a proibição**, e a razão é de realidade operacional: em
instituição pequena a mesma pessoa acumula os dois papéis, e proibir
simplesmente impediria o sistema de representar a instituição como ela é.

A contrapartida é decisão de produto, não detalhe de implementação:

- **A auditoria de cada avaliação registra se o avaliador era também o
  coordenador daquele curso no instante do ato.** No instante do ato, não hoje —
  a designação muda, e a pergunta que interessa depois é quem era o quê quando a
  avaliação aconteceu.
- **O relatório sinaliza essas avaliações e permite filtrá-las.**

O resultado é que a coincidência fica **visível e auditável** em vez de
impossível. Quem precisar exigir segregação — uma comissão avaliadora, por
exemplo — consegue enxergar exatamente onde ela não houve, que é mais do que uma
proibição no cadastro entregaria.

---

## Para quem (atores principais)

| Ator | O que faz no sistema |
|---|---|
| **Administrador do Sistema** | Administra a **plataforma**. Cadastra, edita e inativa **instituições**, cria o **primeiro Pesquisador Institucional** de cada uma, gerencia **outros administradores**, e mantém o **catálogo comum de indicadores do INEP**. **Não pertence a nenhuma instituição e não vê o conteúdo de nenhuma:** não alcança metas, cursos, planos, entregas nem os usuários comuns. Do uso do catálogo, vê apenas a **contagem total**, nunca quais instituições usam o quê. É administrador de plataforma, não superusuário que lê tudo — e **isso é decisão de privacidade, não limitação técnica** |
| **Pesquisador Institucional** (PI) | Termo do e-MEC/INEP para quem responde institucionalmente pelos dados da IES. Acumula o papel de **administrador da instituição dele**: cadastra as pessoas, define os perfis de cada uma e **cadastra os cursos**, atribuindo o coordenador. Cria **indicadores próprios** e as **metas** da instituição, monta o **plano de ação de cada curso** e **avalia as entregas**. **Vê tudo da própria instituição** — nunca de outra. **Pode também coordenar curso**, e a coincidência é registrada (ver acima) |
| **Coordenador de Curso** | Coordena **um ou mais cursos**. Enquanto o plano do curso está **vigente**, responde às metas dele **enviando o arquivo comprobatório** de cada entrega e acompanha o próprio desempenho |
| **Professor** | Perfil de base de quem leciona. Hoje não tem tela própria; existe para o vínculo com curso, que o torna coordenador, e para o acúmulo com outros perfis |
| **Aluno** | Perfil padrão de quem é cadastrado sem outra marcação. Hoje não tem tela própria |

**Não há auto-registro.** O **Administrador do Sistema é criado pelo seed do
banco**; ele cadastra as instituições e cria o primeiro Pesquisador
Institucional de cada uma; o PI cadastra as demais pessoas e os cursos.

### Matriz de permissões

Lida como **união**: quem possui mais de um perfil soma as colunas.

| Ação | Adm. do Sistema | Pesq. Institucional | Coord. de Curso | Professor | Aluno |
|---|---|---|---|---|---|
| Cadastrar / editar / inativar **instituição** | ✅ | ❌ | ❌ | ❌ | ❌ |
| Criar o **primeiro PI** de uma instituição | ✅ | ❌ | ❌ | ❌ | ❌ |
| Gerenciar **outros Administradores do Sistema** | ✅ | ❌ | ❌ | ❌ | ❌ |
| Manter o **catálogo do INEP** (indicadores de escopo plataforma) | ✅ | ❌ | ❌ | ❌ | ❌ |
| Criar **indicadores próprios** da instituição | ❌ | ✅ | ❌ | ❌ | ❌ |
| Criar **metas** da instituição, apontando os indicadores | ❌ | ✅ | ❌ | ❌ | ❌ |
| Cadastrar pessoas e definir perfis **na própria instituição** | ❌ | ✅ | ❌ | ❌ | ❌ |
| Cadastrar **cursos** e atribuir o coordenador | ❌ | ✅ | ❌ | ❌ | ❌ |
| Definir **períodos** | ❌ | ✅ | ❌ | ❌ | ❌ |
| Montar, publicar e encerrar o **plano de ação** de um curso | ❌ | ✅ | ❌ | ❌ | ❌ |
| **Copiar um plano em lote** para vários cursos | ❌ | ✅ | ❌ | ❌ | ❌ |
| **Avaliar entregas** (aceitar ou recusar comprovante) | ❌ | ✅ | ❌ | ❌ | ❌ |
| **Enviar comprovante** de entrega | ❌ | ❌ | ✅ | ❌ | ❌ |
| Ver o **desempenho** de todos os cursos da instituição | ❌ | ✅ | ❌ | ❌ | ❌ |
| Ver o **próprio desempenho** (cursos que coordena) | ❌ | ❌ | ✅ | ❌ | ❌ |
| Ver **conteúdo** de outra instituição | ❌ | ❌ | ❌ | ❌ | ❌ |

**Quem acumula PI e Coordenador soma as duas colunas** — e portanto pode avaliar
entrega de curso que coordena. É permitido, e cada avaliação nessa condição fica
marcada na auditoria e no relatório.

### Visibilidade — duas fronteiras, ambas no use case

**Fronteira externa: a instituição**, com a exceção do catálogo de plataforma
descrita acima.

**Fronteira interna**, dentro da instituição:

| Recurso | Quem possui o perfil de PI | Quem possui o de Coordenador |
|---|---|---|
| **Indicador** | Os do catálogo do INEP e os próprios da instituição, em edição dos próprios | os mesmos, em leitura |
| **Meta** (catálogo da instituição) | Todas as da instituição | em leitura, pelo que aparece nos planos dos cursos dele |
| **Período** | Todos os da instituição | em leitura |
| **Plano, item do plano e entrega** | Todos os da instituição | **Os dos cursos que coordena** |

Consequências para o `arquiteto`:

- A consulta de listagem aplica os recortes na própria query.
- O acesso a um recurso **individual** por identificador também verifica as
  regras — sem isso, trocar o id na URL expõe dado de terceiro (IDOR).
- Vale igualmente para **arquivos**: baixar o comprovante de uma entrega ou o
  documento de um plano passa pelas mesmas verificações. Link de download não
  pode ser caminho público do MinIO.
- **Quem perde a coordenação de um curso perde o acesso aos planos e entregas
  daquele curso.** O recorte é pelo vínculo atual, não por quem enviou o quê.

---

## Escopo da primeira versão (v1)

Cinco funcionalidades, encadeadas. O detalhe de cada uma está na spec dela; aqui
fica o que atravessa.

### 1. Instituições e autenticação — `autenticacao-usuarios`

- **Instituição** como entidade: nome, sigla, código e-MEC (opcional, porque IES
  em credenciamento ainda não o tem) e situação ativa/inativa. **Sigla e código
  e-MEC são únicos entre todas as instituições**, ativas ou inativas. Cadastrada,
  editada e inativada **somente pelo Administrador do Sistema**. **Instituição
  não é excluída, apenas inativada.**
- **Combo de instituição na tela de login**, carregado de **rota pública** — a
  única rota pública de leitura do sistema.
- Cadastro de usuário com **nome, e-mail, senha e conjunto de perfis**, sempre
  dentro de uma instituição (exceto o Administrador do Sistema).
- Login por **instituição + e-mail + senha**. JWT em cookie `HttpOnly`.
- Senha sempre com hash (`argon2id` recomendado), nunca reversível, nunca em log.
- **Não há política de senha nem limite de tentativas de login** — decisão
  explícita do dono, com o risco registrado na spec.
- **Troca de senha obrigatória no primeiro acesso.**
- Senha inicial e esquecimento de senha: ver **premissa P1**.

### 2. Indicador e Meta — os dois catálogos — `indicadores`

- **Indicador em dois escopos:** o **catálogo do INEP, comum à instalação**,
  mantido pelo Administrador do Sistema, e os **indicadores próprios de cada
  instituição**. O indicador de escopo plataforma carrega a **referência do
  instrumento** (dimensão e número).
- **Meta: catálogo da instituição**, apontando para **um ou mais indicadores** —
  pelo menos um. O limite superior existe para que "esta meta atende tudo" não
  vire prática.
- **Uma entrega atende todos os indicadores da meta de uma vez, e nunca é
  contada duas vezes.** É o que evita que a instituição duplique a meta para
  contar em dois indicadores e passe a exigir o dobro de entregas.
- A cadeia de rastreabilidade — `plano → item → meta → indicador(es) →
  referência do instrumento` — é o que responde **"quais indicadores do
  instrumento este curso está atendendo"**.

### 3. Cursos — `cursos`

- **Pertence à instituição.** **Tem um único coordenador**, e um Coordenador
  pode coordenar **vários cursos**.
- **Cadastrado pelo Pesquisador Institucional**, que atribui o coordenador por
  designação.
- O vínculo de coordenação **acrescenta o perfil** `coordenador_curso` a quem o
  recebe, e retirá-lo do último curso **remove o perfil**, preservando os demais.

### 4. Período e Plano de Ação Curso/Coordenador — `plano-acao`

- **Período** pertence a esta feature.
- **O plano é de um curso e de um período** — um plano por par (curso, período).
  Contém descrição, objetivo geral, resultados esperados, alinhamento com o
  **PDI** e o **PPC**, e o órgão de aprovação.
- **Os itens do plano** são o coração: cada item é uma **meta do catálogo com a
  quantidade exigida daquele curso**. **A quantidade é do item, nunca da meta**,
  nunca dividida entre cursos e nunca multiplicada pelo número de indicadores.
- **Situação: rascunho, vigente e encerrado.** **Só a partir de vigente o
  coordenador é cobrado** — é a linha que separa o que o PI ainda está montando
  do que já é compromisso.
- O sistema **gera o documento do plano em `.docx`**.
- O plano pode ser **copiado em lote para vários cursos**, que é o que poupa o
  trabalho do PI quando a mesma exigência vale para muitos cursos.

### 5. Entrega, avaliação e desempenho — `metas-coordenacao`

- O **Coordenador responde enviando arquivo comprobatório** (MinIO) — na v1 o
  comprovante é sempre arquivo anexado.
- O **PI avalia a entrega**, aceitando ou recusando. **A recusa notifica o
  coordenador por e-mail** e por aviso dentro da aplicação.
- **Cada avaliação registra se o avaliador era também o coordenador daquele
  curso no instante do ato**, e o relatório sinaliza e permite filtrar.
- Há **telas de desempenho**: a do PI, com todos os cursos da instituição, e a
  do coordenador, com os cursos que ele coordena.

### Avisos dentro da aplicação

Dois **badges** no shell autenticado, cada um visível a quem interessa:

| Badge | Quem vê | O que conta |
|---|---|---|
| **Entregas pendentes de avaliação** | quem possui o perfil de PI | comprovantes enviados aguardando o parecer dele |
| **Recusas não vistas** | quem possui o perfil de Coordenador | entregas recusadas que ele ainda não abriu |

Ambos contam apenas o que pertence à **instituição da sessão**.

---

## Funcionalidades identificadas (v1)

| # | Funcionalidade | Pasta em `specs/` | Depende de | Fase |
|---|---|---|---|---|
| 1 | **Instituições**, autenticação e cadastro de usuários | `autenticacao-usuarios` | — | em implementação |
| 2 | **Indicador e Meta** — os dois catálogos | `indicadores` | 1 | spec fechada, em validação |
| 3 | **Cursos** | `cursos` | 1, 2 | spec fechada, em validação |
| 4 | **Período e Plano de Ação Curso/Coordenador** | `plano-acao` | 3 | spec fechada, em validação |
| 5 | **Entrega, avaliação e desempenho** | `metas-coordenacao` | 4 | spec fechada, em validação |

**Ordem de implementação — cadeia única, sem paralelismo:**

```
autenticacao-usuarios ──→ indicadores ──→ cursos ──→ plano-acao ──→ metas-coordenacao
```

Sem identidade não há coordenador; sem indicador não há meta; sem curso não há a
quem cobrar; sem plano vigente não há o que entregar.

### Massa de seed compartilhada

O seed atravessa as features, e por isso o que ele precisa conter está
registrado aqui — os vínculos concretos ficam na spec de cada uma:

- **O catálogo de indicadores do INEP**, criado pelo **seed da plataforma**, sem
  instituição. É o que permite que uma instalação nova já tenha o instrumento
  disponível sem ninguém digitar nada.
- **Beatriz Andrade como Pesquisadora Institucional e Professora, com designação
  vigente em um curso**, e **ao menos uma entrega avaliada por ela naquele
  curso**. Sem esses dois fatos juntos, a **marca de coincidência** entre
  avaliador e coordenador e a métrica correspondente **não são observáveis** —
  não há como testar nem demonstrar o controle que substituiu a proibição de
  acúmulo.

---

## Menu hierárquico (AppShell)

**O menu é montado pela união dos perfis do usuário**, a partir de uma
configuração declarativa — sem repetir grupo nem item quando dois perfis
oferecem o mesmo.

**Administrador do Sistema:**

```
┌────────────────────────────────────┐
│ ▾ Sistema                          │
│    · Instituições                  │
│    · Administradores do Sistema    │
│    · Indicadores do INEP           │  ← primeira tela de conteúdo dele
└────────────────────────────────────┘
```

A partir da listagem de instituições, uma ação por linha leva aos
**Pesquisadores Institucionais** daquela instituição — não é item de menu,
porque só existe no contexto de uma instituição escolhida.

**Quem possui o perfil de Pesquisador Institucional** vê os itens das features
2, 4 e 5 reunidos em um grupo de metas, mais o grupo Administração com Usuários
e Cursos.

**Quem possui o perfil de Coordenador de Curso** vê as próprias metas, com o
badge de recusas não vistas, e o próprio desempenho.

**Quem possui ambos** vê o grupo de metas **uma única vez**, com os itens
reunidos, mais o grupo Administração.

**Quem possui apenas Professor ou Aluno** não tem itens de menu nesta versão.

- **"Indicadores do INEP" no grupo Sistema** é a primeira tela de conteúdo do
  Administrador do Sistema. O PI tem a tela dele de indicadores, com o catálogo
  do INEP em leitura e os próprios em edição — são telas diferentes para
  necessidades diferentes.
- **Cursos fica em Administração**, ao lado de Usuários: é cadastro de apoio
  feito pelo PI, e agrupá-lo com as metas sugeriria que só serve para isso.
- **Nenhum item de menu é criado para tela que ainda não existe.**
- **O rótulo exato, a rota e o agrupamento fino dos itens das features 2 a 5 são
  da spec de cada uma** — aqui fica a estrutura e a visibilidade por perfil,
  para que as specs não precisem repetir a regra de montagem do menu.

---

## Restrições gerais do produto

- **Isolamento entre instituições é inegociável**, com a única exceção do
  catálogo de indicadores de escopo plataforma, que é incapaz por construção de
  vazar dado de terceiro. **A Meta não é compartilhada.**
- **Consultas perguntam "possui o perfil X", nunca "é X".**
- **Idioma:** português do Brasil. Datas em `dd/MM/aaaa`, fuso
  `America/Sao_Paulo`.
- **Sem funcionamento offline.**
- **Autorização por recurso desde o início**, avaliada no use case, não na tela.
- **Nenhum arquivo em disco local do container** — comprovantes e documentos de
  plano vão para o armazenamento externo. O sistema roda em múltiplas réplicas
  desde o início.
- **Auditoria** é parte do produto:
  - a trilha de quem entregou e de quem avaliou é o que dá credibilidade ao
    relatório;
  - **cada avaliação registra se o avaliador era também o coordenador daquele
    curso no instante do ato** — é o controle que substituiu a proibição de
    acúmulo, e sem ele a decisão do dono ficaria sem contrapartida;
  - **a mudança do conjunto de perfis registra o conjunto anterior e o novo,
    completos** — perder o que saiu é perder metade da prova de uma escalada de
    privilégio.
- **Falha de e-mail nunca derruba a operação de negócio que a disparou.** A
  recusa de uma entrega é gravada mesmo que a notificação não saia; o aviso
  dentro da aplicação é o canal primário, e o e-mail complementa.
- **Acessibilidade:** navegação por teclado e contraste conforme WCAG 2.2 AA.
- **Responsividade:** celular, tablet e desktop.

---

## LGPD — classificação dos dados pessoais tratados na v1

**A instituição é a fronteira do tratamento.** A mesma pessoa com conta em duas
instituições tem **dois conjuntos de dados independentes**, e um pedido de
titular é respondido **por instituição**.

| Campo | Faixa | Finalidade | Retenção |
|---|---|---|---|
| `nome` | comum | identificar a pessoa nas telas, nos registros de entrega e avaliação e no documento do plano | ver Q2 |
| `email` | comum | credencial de acesso, identificação e **destino da notificação de recusa** | ver Q2 |
| `senha` (hash) | credencial — nunca exposta, nunca em log, nunca reversível | autenticação | enquanto a conta existir |
| `perfis` (conjunto) | comum | autorização | ver Q2 |
| vínculo com a instituição | comum | delimitar a fronteira de acesso ao dado | ver Q2 |
| designação de coordenação de curso | comum | decidir quais planos a pessoa acessa, por quais responde, e **marcar a coincidência entre avaliador e coordenador** | ver Q2 |
| autoria e data das entregas e avaliações | comum | trilha de quem entregou o quê e de quem aceitou ou recusou | ver Q2 |

- **Não há dado sensível** (art. 5º, II) nem **identificador forte** (CPF, RG).
  Pela regra de minimização, nenhum será adicionado sem finalidade declarada.
- **O catálogo do INEP não contém dado pessoal** — é texto do instrumento. É o
  que torna a exceção ao isolamento defensável também do ponto de vista de
  proteção de dados, e não só de arquitetura.
- **Atenção ao conteúdo dos comprovantes:** o arquivo anexado a uma entrega é
  escolhido pelo coordenador e **pode conter dado pessoal de terceiros** que o
  sistema não classifica nem controla. Precisa ser tratado na spec de
  `metas-coordenacao` — no mínimo, orientação na tela de envio e verificação de
  autorização no download.
- **O documento `.docx` do plano** carrega nomes e é gerado e armazenado pelo
  sistema — o download passa pela mesma verificação de autorização dos demais
  arquivos.
- **O e-mail de notificação é transferência de dado para fora do sistema.** O
  conteúdo deve dizer o mínimo — que há uma entrega recusada e onde vê-la —,
  **nunca o comprovante, o motivo detalhado nem dado de terceiro**.
- **O sistema não registra tentativa de login** (decisão do dono). Consequência
  favorável: não há armazenamento do e-mail tentado nem do endereço de origem de
  quem falha ao entrar.
- **Travessias da fronteira institucional, ambas por desenho:** o Administrador
  do Sistema trata nome e e-mail dos **Pesquisadores Institucionais** que cria, e
  vê **a contagem total** de uso de um indicador, sem identificar instituição.
- **Atenção ao perfil Aluno:** quem o possui pode ser **menor de 18 anos**. O
  sistema não coleta data de nascimento. Ver **Q8**.
- **Minimização:** o cadastro de pessoa tem quatro campos (nome, e-mail, senha,
  perfis) mais o vínculo institucional.
- **Log e auditoria nunca contêm senha nem token.**
- **Ambiente não-produtivo nunca recebe dado real.** O seed usa instituições e
  pessoas fictícias plausíveis, e **o Mailpit garante que nenhum e-mail de
  desenvolvimento chegue a uma caixa real**.
- **Base legal:** depende de cada instituição, e **públicas e privadas podem
  coexistir na mesma instalação** — o sistema não declara base legal única. IES
  privada: legítimo interesse (art. 7º, IX) somado a obrigação regulatória
  (art. 7º, II). IES pública: execução de política pública (art. 7º, III).
- **Prazo de retenção não definido** — ver Q2. **Direito de eliminação** — Q3.

---

## Decisões técnicas registradas

A stack completa está em `project.config.md`.

- **Backend Go + Gin com sqlx**; **frontend Next.js + shadcn/ui**;
  **PostgreSQL** em produção **e em desenvolvimento**.
- **Isolamento multi-institucional por coluna, banco único**, com o filtro
  centralizado no adapter de persistência. **O filtro é "instituição da sessão ou
  escopo de plataforma"**, e a segunda parte só alcança linhas sem instituição.
- **Perfis em tabela de vínculo usuário–perfil**, não em coluna. O conjunto é
  lido a cada requisição autenticada — é a consulta mais frequente do sistema — e
  as invariantes de "último detentor" exigem verificação e escrita na mesma
  transação, com bloqueio adequado.
- **Armazenamento de arquivos: MinIO** em dev, MinIO ou S3 em produção. Guarda
  **os comprovantes das entregas e os documentos `.docx` dos planos**. O download
  passa por rota autenticada da aplicação — nunca URL pública do bucket.
- **Geração de documento `.docx`** para o plano de ação — a biblioteca e o
  desenho do adapter ficam no `design.md` de `plano-acao`. A regra que já
  vale é a arquitetural: adapter de saída em `/adapter` atrás de interface em
  `/port`.
- **Envio de e-mail: sim**, a partir de `metas-coordenacao`. **Mailpit** em
  desenvolvimento, que **captura a mensagem sem entregar para fora**; **SMTP**
  corporativo em produção. Adapter de saída atrás de interface em `/port`, com
  circuit breaker.
  - **Ponto de atenção para o `arquiteto`:** gravar a recusa no banco e enviar o
    e-mail é um **dual write** (CLAUDE.md). Sem mensageria, a escolha entre
    Outbox e reconciliação por job fica no `design.md` de `metas-coordenacao`.
- **Sem mensageria, sem IA/LLM, sem dados geográficos e sem editor de texto
  rico** — nenhuma feature do escopo atual os justifica.
- **API:** prefixo `/api/v1`, erro no formato
  `{"error": {"code": "...", "message": "..."}}`.

### Estratégia de testes (decisão de custo do dono do produto)

Nas fases iniciais constrói-se apenas o **mínimo necessário** de testes
automatizados; o restante é **anotado** e vira suíte na fase de Release.

- **Coberto desde já:** o que é fronteira de segurança, isolamento ou integridade
  de dado — isolamento entre instituições (**incluindo a exceção do catálogo, que
  precisa provar que escopo de plataforma não vaza nem alcança dado de
  instituição**), invariantes com condição de corrida, união de permissões,
  matriz de autorização (incluindo o 404 de outra instituição), concorrência
  otimista e unicidade, mais o smoke dos endpoints. Testes de use case escritos
  **antes** do código.
- **Adiado:** caminho feliz de CRUD, paginação, ordenação, persistência de
  filtro, validação de formato, estados de tela, microcópia e testes de
  componente.
- **Onde o adiado fica registrado:** `specs/<feature>/testes-pendentes.md`, uma
  linha por cenário com prioridade — insumo direto da suíte E2E da Release.
- **Consequência declarada:** até a Release, uma regressão nos caminhos adiados
  não é detectada automaticamente.
- **Critério de aceitação:** cada cenário da spec está **ou** coberto por teste
  automatizado que passa, **ou** listado em `testes-pendentes.md` e validado
  manualmente no smoke do dono — nunca em nenhum dos dois, nunca nos dois.
- **E2E com Playwright não é escrito na v1.** Entra na fase de Release, tendo
  login e isolamento como os dois primeiros fluxos.

---

## Fora de escopo (explicitamente, para não reaparecer depois)

- **Geração do PPC completo** — é a visão de longo prazo, não a v1.
- **Atas de reunião com assinatura eletrônica** — retiradas do escopo em
  28/09/2026. Ver o roadmap.
- **Proibir o acúmulo de Pesquisador Institucional e Coordenador de Curso**
  (segregação de funções). Proposto e **recusado pelo dono**: a coincidência é
  marcada na auditoria e no relatório, não impedida no cadastro.
- **Compartilhar a Meta entre instituições.** A exceção ao isolamento alcança o
  indicador de escopo plataforma e para aí.
- **Qualquer envio de e-mail além da notificação de recusa de entrega.**
- **Auto-registro de usuário** e login federado (SSO, GOV.BR, Google).
- **Conta única vinculada a várias instituições, com troca de contexto na
  sessão.** Considerado e recusado.
- **Subdomínio por instituição.** Considerado e recusado.
- **Política de senha e limite de tentativas de login** — decisão do dono, com
  risco registrado.
- **Exclusão de instituição.** Instituição é inativada, nunca excluída.
- **Curso com mais de um coordenador**, e coordenação compartilhada ou suplente.
- **Comprovante de entrega que não seja arquivo anexado.**
- **Acúmulo do perfil Administrador do Sistema com qualquer outro** — razão
  estrutural, ver a seção de perfis.
- **Qualquer outra exceção ao isolamento além do catálogo de indicadores.**
- **Perfis configuráveis em runtime.** Os cinco são fixos.
- **Mensageria, processamento assíncrono e exportação em lote.**
- **Qualquer capacidade de IA/LLM.** **Dados geográficos / mapas.**

---

## Roadmap (não implementar agora)

| Item | Descrição | Pré-requisito |
|---|---|---|
| **Atas de reunião com assinatura eletrônica** | Modelo de ata reutilizável, criação de ata a partir dele e aceite eletrônico auditado, para servir de evidência documental ao INEP. **Retirado do escopo em 28/09/2026, a reavaliar** — registrado para que a decisão não se perca e a ideia não volte como nova | decisão do dono do produto |
| **Catálogo estruturado do instrumento do INEP** | Hoje a referência do instrumento é texto livre no indicador de escopo plataforma. Estruturar em instrumento, versão, dimensão e número permitiria navegar e comparar | volume de instituições que justifique |
| **Conta única multi-institucional com troca de contexto** | Uma credencial, várias instituições, com seletor na sessão. **Recusado agora** | demanda real de quem acha duas contas insuficiente |
| **Subdomínio por instituição** | `fsa.basis-avalia.br` em vez de combo. **Recusado agora** | decisão de DNS e certificado |
| **Exclusão de instituição** | Bloqueada enquanto houver usuário, curso, plano ou meta vinculada | pedido concreto; hoje inativar resolve |
| **Login com Google (OAuth)** e **sincronização do Google Workspace** | Dois itens distintos. O modelo de identidade da v1 já os acomoda | decisão do dono |
| **Política de senha e limite de tentativas** | Os controles recusados, com o risco hoje aceito | pedido do dono |
| **Geração do PPC completo** | Montar o PPC a partir dos instrumentos do INEP e das evidências acumuladas | modelagem dos instrumentos |
| **Suíte E2E e os testes adiados** | Todo o conteúdo de `testes-pendentes.md` vira suíte automatizada | declaração da primeira versão estável |
| **Oportunidades de IA** | Acionar o `ai-consultor` para avaliar oportunidades sobre o escopo atual | v1 concluída |

---

## Premissas declaradas

| # | Premissa | Onde confirmar |
|---|---|---|
| **P1** | **Senha inicial e esquecimento:** quem administra define a senha no cadastro, e a pessoa a altera no primeiro acesso, obrigatoriamente. **Não há recuperação por e-mail** — nem depois que o envio de e-mail passou a existir, porque foi habilitado para um caso específico. Esquecimento é resolvido por quem administra: o PI para as pessoas da instituição dele, o Administrador do Sistema para os PIs | `autenticacao-usuarios` — **confirmada** |

As premissas de cada feature estão na spec dela: multi-institucionalidade em
`autenticacao-usuarios` (seção 1), catálogos em `indicadores` (PI-1 a PI-5),
plano em `plano-acao` (PP-1 a PP-9).

---

## Questões abertas (negócio)

| # | Questão | Resolver na spec de |
|---|---|---|
| **Q2** | **Por quanto tempo os dados ficam** e o que acontece depois (descarte ou anonimização)? Vale para o cadastro de pessoa, para os comprovantes e para os documentos de plano. Sem prazo declarado, o dado fica para sempre — o pior padrão possível | `autenticacao-usuarios`, `metas-coordenacao` |
| **Q3** | **Pedido de eliminação de dados de um titular que já enviou ou avaliou comprovante:** anonimizar o nome mantendo a trilha íntegra, ou outro tratamento? A resposta é **por instituição** | `metas-coordenacao` |
| **Q5** | **Volume esperado:** quantas instituições, cursos, indicadores, metas e usuários no primeiro ano e em dois anos? | `metas-coordenacao` (e insumo para o `arquiteto`) |
| **Q6** | **O que é inaceitável acontecer?** A leitura é: (a) perder um comprovante enviado, (b) uma entrega aparecer avaliada por quem não a avaliou, (c) alguém ver plano ou entrega que não lhe compete, (d) **dado de uma instituição aparecer para outra**. Confirma? | `metas-coordenacao` |
| **Q7** | **Como a cobrança de metas é feita hoje**, antes do sistema? É onde aparecem as regras que ninguém verbaliza | `cursos`, `metas-coordenacao` |
| **Q8** | **Dado de quem tem o perfil Aluno e é menor de 18 anos:** confirmar a recomendação de manter a não coleta de data de nascimento | `autenticacao-usuarios` |

A questão da **base legal** deixou de ser bloqueante: instituições públicas e
privadas podem coexistir na mesma instalação, as duas hipóteses estão na seção de
LGPD, e nenhuma linha de código depende da escolha.

---

## Imagens/referências fornecidas

Nenhuma imagem, mockup ou documento de referência foi fornecido. Os
**instrumentos de avaliação do INEP** são a base conceitual do catálogo de
indicadores, mas o documento em si não foi anexado — o conteúdo concreto do
catálogo entra pelo seed e pela tela do Administrador do Sistema, com a
referência do instrumento registrada como texto livre em cada indicador de
escopo plataforma.
