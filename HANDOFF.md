# gotextforge — Handoff

> Documento de contexto para qualquer sessão futura (nova conversa de IA ou retomada depois de pausa). **Leia este arquivo antes de tudo**, depois o `ROADMAP.md`.

## Estado atual (atualizado em 2026-11-12)

**MS0 e MS1 concluídos.** O editor abre um arquivo pela linha de comando e mostra o conteúdo no terminal (Bubble Tea v2). O scroll funciona com `j`/`k`/↑/↓/PgUp/PgDn/`g`/`G`/Home/End, e Ctrl+c sai. Testado à mão com `runtime/proc.go` (8 mil linhas) e `ssa/rewriteAMD64.go` (110 mil linhas) da stdlib.

**Próximo:** MS2, edição com modos (`i`, Esc, digitar, salvar, undo).

### Estrutura real (difere do ROADMAP, que ainda mostra `core/` na raiz)

```
cmd/gotextforge/main.go     lê os.Args[1] → buffer.Load → editor.New → tui.Run; erros saem antes do TUI
internal/core/input/        Event{Kind, Key, Width, Height}, Key{Code, Rune, Mod} (comparável, vira chave de map)
internal/core/render/       Frame{Lines, Cursor ScreenPos, Status}, StyledLine = []Span, Style, Color{Kind ANSI|RGB}
internal/core/buffer/       [][]rune somente leitura; Load(io.Reader); LineEnding (LF/CRLF); HasBOM; ErrUTF16
internal/core/view/         Viewport{Top, Height} com clamp; FormatLine (tabs, corte na largura, controles → ^X / <xx>)
internal/core/editor/       junta buffer + viewport; HandleEvent(Event); Frame() render.Frame
internal/frontend/tui/      TranslateKey (tea → input.Key), RenderFrame (Frame → string via Lip Gloss), model
```

O core (`internal/core/...`) não importa nada de UI. Isso é verificado em toda revisão.

### Decisões tomadas durante MS0–MS1 (complementam D1–D8)

- **`internal/`**: o core fica em `internal/core/` e o frontend em `internal/frontend/`, para que outros módulos não importem.
- **API síncrona por enquanto**: o frontend chama `editor.HandleEvent` e `editor.Frame()` diretamente. O event loop próprio do roadmap (goroutine + channels) vai embrulhar essa API depois, sem mudar nenhum tipo.
- **`Color`** é uma struct com `ColorKind` (Default/ANSI/RGB). O zero value é a cor padrão. `ANSI(n)` aceita a paleta inteira de 0 a 255.
- **Normalização de teclas**: em caracteres, o Shift nunca é marcado (`A` vira `Rune 'A'`). Em teclas especiais, sim (`<S-Tab>`). Ctrl+letra usa minúscula (`'a'` + `ModCtrl`). `Key.String()` usa a notação do vim (`<C-M-Up>`, `<CR>`, `<Space>`).
- **Frame do tamanho da tela**: o core devolve exatamente `altura-1` linhas (completa com `~`) mais o status. O frontend não faz layout.
- **Tabstop 4**, com a tab indo até a próxima parada. Por enquanto, 1 rune ocupa 1 célula: o wcwidth para CJK é uma **limitação conhecida**, ainda não feita.
- **Codificação**: UTF-8 é o padrão. O BOM de UTF-8 é removido e lembrado. UTF-16 é recusado com `ErrUTF16`.
- **Teclas de scroll no `editor.handleKey` são andaime**: a engine vim do MS4 substitui essa parte.

### Medições (benchmarks em `*_bench_test.go`, `go test -run '^$' -bench=. -benchmem ./...`)

- `Frame()` leva cerca de 50 µs com 1 mil, 10 mil ou 100 mil linhas: a virtualização (D8) está comprovada.
- `RenderFrame` com 50 linhas: cerca de 0,06 ms com 1 span por linha, 1,2 ms com 10 e 3,8 ms com 30. **Esse é o gargalo esperado no MS3 (highlight)**: guardar o `lipgloss.Style` pronto por `render.Style`, já que o `Style` é comparável e pode ser chave de map. Só otimizar quando o número mandar.
- `buffer.Load` de 100 mil linhas leva cerca de 39 ms (acontece uma vez, ao abrir).

## Objetivo e restrições do projeto

- **O que:** editor de texto/IDE para Go com dois frontends — terminal (estilo LazyVim/Neovim, fase 1) e desktop Wails (fase 2) — sobre um núcleo Go único sem dependência de UI.
- **Por quê:** produto + aprendizado (o relatório de pesquisa recomenda este projeto sobre o alternativo de tarefas distribuídas).
- **Ritmo:** ~5h/semana, sem prazo.
- **Perfil de quem constrói:** domina Go e Java; já usa Wails no projeto `newpostman` (C:\Projetos\WAILS\newpostman); prefere que decisões de arquitetura sejam tomadas **antes** do código para evitar refatoração.
- **Modo de colaboração:** o humano constrói do zero; a IA guia (objetivo, atenção, critério de aceite) e revisa código quando colado. Não escrever código do projeto por ele, a menos que peça.
- **Como as sessões funcionam:** um passo pequeno por vez, com o conceito explicado antes. O humano escreve "verifica" e a IA inspeciona o repositório (gofmt, vet, testes, imports do core), roda probes quando precisa e revisa. O humano é leigo em editores, TUI e LSP, então a explicação deve ser didática.
- **Sem comentários no código, de propósito:** o humano anota em outro lugar. Não cobrar doc comments nas revisões.

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

1. ~~Milestone 0~~ ✅ e ~~Milestone 1~~ ✅.
2. **Milestone 2**: cursor com posição no arquivo (linha + rune, D3) convertida para `ScreenPos`; modos normal e insert; digitar; salvar respeitando LineEnding e BOM; undo.
3. Seguir a tabela de milestones do ROADMAP.md; cada sessão começar pedindo objetivo + critério de aceite.
4. Pendente (do humano): atualizar a árvore do `ROADMAP.md` para `internal/`.

## Armadilhas já conhecidas (não redescobrir)

- LSP exige posições **UTF-16** — mapear só no cliente LSP (D3).
- Quebra de linha/width precisa do mesmo `wcwrite` no TUI e no desktop (contrato).
- Renderizar linha a linha em DOM = armadilha de performance do Wails; canvas único com virtualização.
- Vim engine com pending state desde o início — keymap de frontend nunca deve saber o que é operador.
- Se um milestone passar sem algo rodável visível, recortar escopo, não adiar data.
- **Windows/CRLF**: o `.gitattributes` (`* text=auto eol=lf`) força LF. Se o `gofmt -l` acusar arquivos que parecem formatados, é o editor salvando em CRLF.
- **O Windows PowerShell 5.1 grava UTF-16 com `>`/`echo`** (foi assim que o README nasceu quebrado). Criar arquivos pelo editor.
- **`String()` com receiver de ponteiro não é chamado pelo `fmt` com valores**: use receiver de valor em tipos pequenos.
- **`bufio.Scanner` já descarta o newline final e o `\r`**: não aparar à mão (isso já engoliu linhas em branco no fim do arquivo).
- **Conteúdo do arquivo nunca vai cru para o terminal**: `FormatLine` neutraliza os caracteres de controle (ESC, `\r`, C1).
- **Risco no MS2**: um arquivo com bytes UTF-8 inválidos (sem BOM) abre com `�`, e ao salvar o original se perde. Decidir no MS2 (provavelmente abrir como somente leitura).
