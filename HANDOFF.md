# gotextforge — Handoff

> Documento de contexto para qualquer sessão futura (nova conversa de IA ou retomada depois de pausa). **Leia este arquivo antes de tudo**, depois o `ROADMAP.md`.

## Estado atual

**Nada foi codificado ainda.** O projeto existe apenas como decisão e documentação:

- `ROADMAP.md` — plano completo, contratos e decisões travadas.
- Este `HANDOFF.md` — contexto e onde estamos.
- Pesquisa de mercado/técnica concluída (relatório abaixo).
- Próximo passo concreto: **milestone 0** (esqueleto do repo `go mod init` + tipos do contrato `Event`/`Frame`/`StyledLine`/`Style`), depois **milestone 1** (viewer).

## Objetivo e restrições do projeto

- **O que:** editor de texto/IDE para Go com dois frontends — terminal (estilo LazyVim/Neovim, fase 1) e desktop Wails (fase 2) — sobre um núcleo Go único sem dependência de UI.
- **Por quê:** produto + aprendizado (o relatório de pesquisa recomenda este projeto sobre o alternativo de tarefas distribuídas).
- **Ritmo:** ~5h/semana, sem prazo.
- **Perfil de quem constrói:** domina Go e Java; já usa Wails no projeto `newpostman` (C:\Projetos\WAILS\newpostman); prefere que decisões de arquitetura sejam tomadas **antes** do código para evitar refatoração.
- **Modo de colaboração:** o humano constrói do zero; a IA guia (objetivo, atenção, critério de aceite) e revisa código quando colado. Não escrever código do projeto por ele, a menos que peça.

## Contexto de pesquisa (feita, não refazer)

- Relatório: `C:\Projetos\WAILS\newpostman\reports\Editor Go ou tarefa distribuida.md`
- Notas brutas: `C:\Projetos\WAILS\newpostman\research_notes\Editor Go ou tarefa distribuida\`
- Conclusão principal: editor Go > tarefas distribuídas para este perfil/ritmo; o TUI-first veio depois como requisito explícito do usuário ("quero também no terminal, tal qual lazyvim/neovim").

## Resumo das decisões (detalhe no ROADMAP.md)

1. **Começar como Go puro, não `wails init`** — Wails entra depois como dependência de `frontend/desktop/`.
2. **TUI primeiro (Bubble Tea v2), desktop depois (Wails v3 + canvas).**
3. **Core com contrato de eventos/frames** (entrada `Event`, saída `Frame{[]StyledLine}`) — a salvaguarda contra refatoração e o que permite os dois frontends.
4. **Vim modal no core** como linguagem de comandos com pending state.
5. Buffer `[][]rune` com gap buffer por linha; posições rune+linha, UTF-16 só na fronteira LSP; highlight via `go/scanner`; LSP client próprio; virtualização de linhas desde o milestone 3.

## Ferramentas e referências validadas (2026)

- **Bubble Tea v2** — lançado em 2026 (charm.land/blog/v2), ativo. Stack: Bubble Tea + Lip Gloss + Bubbles (v2).
- **Wails v3.0.0** — estável desde 26/10/2026; recupera crash do WebView2. Issue conhecido: lag com HTML/CSS pesado (wails#4155) — por isso a área de texto é `<canvas>`, não DOM.
- **Fyne** — descartado (TextGrid com lag/vazamento e não editável: fyne#2874, fyne#3994).
- **Gio** — plano B do plano B; gvcode (github.com/oligo/gvcode) é componente de editor pra Gio, referência se o canvas falhar.
- **Referências de estudo:** `micro` (editor de terminal em Go, tcell) e o tutorial `kilo` (viewsourcecode.org/snaptoken/kilo) para buffer/editor mínimo.

## Próximos passos (ordem)

1. Milestone 0: `go mod init github.com/<user>/gotextforge`, estrutura de pastas do ROADMAP, tipos do contrato.
2. Milestone 1: viewer com Bubble Tea v2 (abre arquivo, renderiza, scroll) — validar com arquivo `.go` de 5k+ linhas.
3. Seguir a tabela de milestones do ROADMAP.md; cada sessão começar pedindo objetivo + critério de aceite.

## Armadilhas já conhecidas (não redescobrir)

- LSP exige posições **UTF-16** — mapear só no cliente LSP (D3).
- Quebra de linha/width precisa do mesmo `wcwrite` no TUI e no desktop (contrato).
- Renderizar linha a linha em DOM = armadilha de performance do Wails; canvas único com virtualização.
- Vim engine com pending state desde o início — keymap de frontend nunca deve saber o que é operador.
- Se um milestone passar sem algo rodável visível, recortar escopo, não adiar data.
