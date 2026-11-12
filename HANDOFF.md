# gotextforge — Handoff

> Documento de contexto para qualquer sessão futura (nova conversa de IA ou retomada depois de pausa). **Leia este arquivo antes de tudo**, depois o `ROADMAP.md`.

## Estado atual (atualizado em 2026-11-12)

**MS0 a MS5 concluídos, mais um visual estilo LazyVim.** O editor abre um arquivo pela linha de comando, ou sem argumento mostra uma tela inicial com logo, no terminal (Bubble Tea v2). Recursos atuais:
- **Vim (MS4, engine em `internal/core/engine`):**
  - counts: `3j`, `2d3w`;
  - motions: `h j k l w W b B e E 0 ^ $ gg G f F t T { }`, setas, PgUp/PgDn;
  - operadores `d c y` com qualquer motion, e `dd cc yy`;
  - text objects: `iw aw iW aW i" a" i' a' i( a( i[ a[ i{ a{ i< a< ip ap`, mais a crase (raw string) em versão `i` e `a`;
  - atalhos `x X D C s S Y`, além de `p P J r`;
  - insert com `i a I A o O`, e undo/redo (`u`, Ctrl+r, com count).
- **Arquivo:** salvar e sair (`:w :q :q! :wq :x`).
- **gopls (MS5):**
  - erros e avisos aparecem enquanto você digita: sublinhado ondulado colorido, mensagem `●` no fim da linha, número da linha colorido e `E2 W1` na status line;
  - `K` abre o hover numa janela flutuante, com o código colorido e o markdown limpo;
  - o completion abre um menu automático ao digitar (no início de uma palavra ou depois de `.`), com Ctrl+n/Ctrl+p ou setas para navegar, Enter/Tab/Ctrl+y para aceitar, Ctrl+e para fechar e Ctrl+Espaço para pedir manualmente;
  - o indicador `gopls…` / `gopls` / `gopls!` na status line mostra o estado do servidor.
- **Visual:** paleta truecolor inspirada no Tokyo Night (assume terminal escuro); números de linha relativos, com o da linha atual em laranja; destaque da linha do cursor; status line estilo lualine (modo colorido, arquivo `[+]`, comando pendente, `go`, progresso, linha:coluna); erros em vermelho; tela inicial com logo em degradê.

O critério do MS2 foi verificado de ponta a ponta: o editor modifica o `main.go` de um módulo real, salva com `:wq`, e o `go build` e o binário rodam.

Arquivos `.go` têm syntax highlight (MS3): palavras-chave, strings, números, comentários, identificadores predeclarados e nomes de função, com comentários de bloco e raw strings que atravessam linhas tratados corretamente.

**Próximo:** MS6, polimento (busca `/`, abas ou buffers, tema configurável) — "algo de postar no r/golang".

### Estrutura real (difere do ROADMAP, que ainda mostra `core/` na raiz)

```
cmd/gotextforge/main.go     [arquivo] → buffer.Load → editor.New → tui.Run; sem argumento abre a tela inicial; erros saem antes do TUI
internal/core/input/        Event{Kind, Key, Width, Height}, Key{Code, Rune, Mod} (comparável, vira chave de map)
internal/core/render/       Frame{Lines, Cursor ScreenPos, CursorShape, Status StyledLine}, StyledLine = []Span (Text(), Width()), Style, Color{Kind ANSI|RGB}
internal/core/buffer/       [][]rune somente leitura; Load(io.Reader); LineEnding (LF/CRLF); HasBOM; ErrUTF16
internal/core/view/         Viewport (clamp, EnsureVisible); layout único → FormatLine (texto) e RenderLine (spans com estilo); ScreenCol
internal/core/syntax/       HighlightLine(linha, estado) via go/scanner; Cache de estado por linha + NextState (atalho exato)
internal/core/storage/      Save atômico (tmp + Sync + Rename)
internal/core/lsp/          cliente LSP próprio (D5): jsonrpc.go (Content-Length, ids, pedidos do servidor), protocol.go (tipos,
                            UTF-16 ↔ rune SÓ aqui (D3), URIs no Windows), client.go (inicia o gopls de forma assíncrona, fila sem bloquear,
                            agrupa mudanças de texto, eventos num canal, shutdown/exit/kill)
internal/core/engine/       vim: Parser (gramática + estado pending, Pending() = showcmd), ApplyMotion, TextObject, ChangeWordTarget
internal/core/editor/       editor.go (estado/eventos), normal.go (executa comandos da engine, registrador, p/J/r),
                            insert.go, command.go (:w :q...), ui.go (Frame: gutter, cursorline, statusline), intro.go, theme.go,
                            lsp.go (LanguageClient, HandleAsync), diagnostics.go, popup.go (hover/janelas), completion.go
internal/frontend/tui/      TranslateKey (tea → input.Key), RenderFrame (Frame → string com SGR ANSI direto), model
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
- **D2 na prática (MS2)**: a edição usa slices simples por linha, ainda sem gap buffer. A API é `Insert(pos, texto) → fimPos` e `Delete(início, fim) → textoRemovido`; o `Delete` devolve o texto justamente para servir de base ao undo. Medido: inserir uma rune numa linha típica leva cerca de 0,4 µs, e numa linha de 100 mil caracteres cerca de 0,26 ms. O gap buffer só entra se algum número mandar.

### Medições (benchmarks em `*_bench_test.go`, `go test -run '^$' -bench=. -benchmem ./...`)

- `Frame()` leva cerca de 50 µs com 1 mil, 10 mil ou 100 mil linhas: a virtualização (D8) está comprovada.
- `RenderFrame` com 50 linhas: com Lip Gloss eram 1,2 ms (10 spans por linha) e 3,8 ms (30 spans). **No MS3 o Lip Gloss foi trocado por SGR ANSI escrito à mão**: agora são 0,03 ms e 0,12 ms. Isso é seguro porque o texto dos spans já vem sanitizado pelo `RenderLine`, e o renderer do Bubble Tea v2 continua adaptando as cores ao perfil do terminal. O Lip Gloss saiu do `go.mod`.
- Quadro completo com highlight (Frame + RenderFrame, 200x50): ~0,6 ms. `Frame()` colorido: ~0,3 ms em qualquer tamanho de arquivo.
- Primeiro `G` num arquivo de 100 mil linhas (o cache calcula os estados de todas as linhas): 273 ms com o `go/scanner` em cada linha; **16 ms** com o `NextState`, que só tokeniza linhas com `/*`, `*/` ou `` ` ``.
- Validação do highlight: o cache de estados foi comparado com o `go/scanner` rodando no arquivo inteiro, em toda a stdlib do Go (8.087 arquivos, ~3 milhões de linhas), com 0 divergências.
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

- **Bubble Tea v2** — lançado em 2026 (charm.land/blog/v2), ativo. Stack hoje: só Bubble Tea v2 (o Lip Gloss foi removido no MS3; veja Medições).
- **Wails v3.0.0** — estável desde 26/10/2026; recupera crash do WebView2. Issue conhecido: lag com HTML/CSS pesado (wails#4155) — por isso a área de texto é `<canvas>`, não DOM.
- **Fyne** — descartado (TextGrid com lag/vazamento e não editável: fyne#2874, fyne#3994).
- **Gio** — plano B do plano B; gvcode (github.com/oligo/gvcode) é componente de editor pra Gio, referência se o canvas falhar.
- **Referências de estudo:** `micro` (editor de terminal em Go, tcell) e o tutorial `kilo` (viewsourcecode.org/snaptoken/kilo) para buffer/editor mínimo.

## Próximos passos (ordem)

1. ~~Milestone 0~~ ✅ e ~~Milestone 1~~ ✅.
2. ~~Milestone 2~~ ✅.
   - Pendências pequenas, boas para o humano fazer: autoindent no Enter e no `o`; `:w outro.go`; abrir um arquivo inexistente como novo.
3. ~~Milestone 3~~ ✅ e ~~Milestone 4~~ ✅. O critério do MS4 (`ciw`, `dap`, `dd` e counts) está coberto em `editor/vim_test.go`, com 66 sequências de teclas.
   - O que ficou fora do MS4 (bom para depois): `.` (repetir), `;`/`,`, modo visual, registradores nomeados (`"a`), marcas, `%`, `ge`, autoindent, count no `i`.
   - ~~Milestone 5~~ ✅. O critério ("erros do compilador em tempo real enquanto digita") está verificado com o gopls real em `editor/gopls_integration_test.go`.
   - O que ficou fora do MS5: `gd` (ir para a definição), renomear, formatar ao salvar, sync incremental (hoje o texto inteiro vai a cada mudança, o que é ok até alguns milhares de linhas) e snippets.
   - **Próximo é o MS6, o polimento.**
4. Seguir a tabela de milestones do ROADMAP.md; cada sessão começar pedindo objetivo + critério de aceite.
5. Pendente (do humano): atualizar a árvore do `ROADMAP.md` para `internal/`.

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
- **UTF-8 inválido (resolvido no MS2)**: o arquivo abre (mostrando `�`), mas o `:w` é bloqueado (`buffer.HasInvalidUTF8`), para não destruir os bytes originais.
- **Salvar = `storage.Save`, atômico**: grava num `.nome.*.tmp` na mesma pasta, faz `Sync` e depois `Rename` por cima do original. Segue symlinks e preserva permissões. Num erro, o original fica intacto.
- **Fidelidade ao salvar**: `buffer.WriteTo` devolve o mesmo LF/CRLF, o mesmo BOM e o mesmo newline final (ou a falta dele) que o arquivo tinha. A ida e volta Load→WriteTo é testada byte a byte. Exceção conhecida: um arquivo com finais de linha **misturados** é normalizado para CRLF se tiver qualquer `\r\n`.
- **O `Status` do `Frame` virou `StyledLine`** (foi `string` até o MS3): o core monta a status line colorida, e o frontend só desenha spans.
- **O tema assume terminal escuro** (cores RGB do Tokyo Night, com fundo de destaque `#292e42`). Num terminal claro, o destaque da linha fica feio; um tema claro ou configurável fica para o MS6.
- **Semântica vim seguida à risca, e testada:** um movimento que falha não mexe na coluna desejada; `cw` age como `ce`; `dw` na última palavra da linha não junta linhas; `i{` em várias linhas vira operação por linhas inteiras; o undo de `ciw` mais o texto digitado é um passo só.
- **Teste aleatório** (`editor/random_test.go`): 80 mil teclas vim aleatórias, com semente fixa, verificando que não há panic, que o cursor é sempre válido e que nenhuma linha passa da largura da tela.
- **Assincronia (MS5):** as goroutines do cliente LSP **nunca** tocam no estado do editor. Elas mandam eventos por `editor.Async()`; o TUI tem um `tea.Cmd` que espera nesse canal e chama `editor.HandleAsync(ev)` na thread da interface. As respostas antigas (hover ou completion com um id que já não é o último) são ignoradas.
- **Encerramento do gopls:** `ed.Close()` no fim do `main` manda `shutdown` e `exit`, e mata o processo depois de 2 s se ele não sair. Conferido no Gerenciador de Tarefas: nenhum gopls órfão sobra dos testes (os processos `gopls` que aparecem rodando são do VS Code).
- **Testes com gopls real** (`lsp/gopls_integration_test.go` e `editor/gopls_integration_test.go`): são pulados com `go test -short ./...` ou sem gopls no PATH. Levam cerca de 2 s.
- **O renderer emite SGR `4:3` (undercurl) e `58;2;r;g;b` (cor do sublinhado)**: o terminal do VS Code e o Windows Terminal suportam; terminais antigos caem para sublinhado simples ou ignoram.
- **Sair é decisão do core**: `editor.ShouldQuit()`. O TUI só obedece; ele não trata mais o Ctrl+c.
- **Undo mora no buffer** (`history.go`): cada `Insert`/`Delete` é registrado. O editor agrupa as mudanças com `BeginChange`/`EndChange`: um comando do modo normal é um passo, e uma sessão inteira de insert (do `i` ao Esc) também é um passo. O undo devolve o cursor para onde ele estava **antes** do comando. O `Modified()` compara o id do topo da pilha com o id salvo, então desfazer até o estado salvo tira o `[+]`.
