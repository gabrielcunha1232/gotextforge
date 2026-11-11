# gotextforge — Roadmap

> Editor de texto/IDE para Go, com frontend de terminal (estilo LazyVim/Neovim) e frontend desktop (Wails), sobre um único núcleo Go sem dependência de UI.
>
> **Ritmo:** ~5h/semana, sem prazo fixo. **Meta:** produto + aprendizado.
> **Regra de ouro:** cada milestone termina com algo rodável e visível.

## Como o projeto começa (decisão travada)

**Projeto Go puro (`go mod init github.com/<user>/gotextforge`), NÃO `wails init`.**

- A fase 1 é TUI — o Wails não tem papel nela.
- O `wails init` nasce com estrutura de webview embutida e cria a tentação de colocar lógica no frontend web.
- O Wails entra depois **como dependência de um subpacote** (`frontend/desktop/`), junto com o spike do canvas. Adicionar o módulo depois é 10 minutos; arrancar suposições de estrutura do Wails depois é refatoração.

## Estrutura do repositório

```
gotextforge/
├── cmd/gotextforge/      # main: CLI, parse de args, escolha de frontend
├── core/                 # ZERO imports de UI (o contrato com o futuro)
│   ├── buffer/           # linhas [][]rune, gap buffer, undo/redo
│   ├── view/             # viewport, scroll, cursor, quebra de linha
│   ├── render/           # Frame, StyledLine, Style
│   ├── engine/           # linguagem de comandos vim (motions/operadores/pending)
│   ├── keymap/           # tecla → Event/Command (sem know-how de vim aqui)
│   ├── lsp/              # cliente gopls: JSON-RPC/stdio, posições UTF-16 só aqui
│   └── syntax/           # highlight via go/scanner (stdlib)
├── frontend/
│   ├── tui/              # Bubble Tea v2: Frame → células ANSI
│   └── desktop/          # Wails v3 + canvas (fase 2, adicionar depois)
└── docs/                 # notas de arquitetura
```

## Contrato core ↔ frontend (a refactor-proofing)

1. **Entrada:** o frontend traduz tecla/mouse/resize para `Event` — tipos abstratos, nunca structs do Bubble Tea nem do DOM no core.
2. **Saída:** o core devolve `Frame{Lines []StyledLine, Cursor, Status}` — orientado a linhas, não células nem pixels.
3. **`Style`** cobre: fg, bg, bold, italic, underline, undercurl. O TUI degrada (undercurl→underline); o canvas desenha fiel.
4. **Largura unicode (wcwidth) vive no core** — o terminal e o desktop compartilham o mesmo cálculo de quebra de linha.
5. **O core roda num event loop próprio:** frontends mandam eventos, recebem frames (modelo Neovim/UI protocol). Permite um 3º frontend no futuro sem tocar no core.

## Decisões travadas (não reabrir sem motivo)

| # | Decisão | Por quê |
|---|---|---|
| D1 | Edição modal vim **no core** como linguagem de comandos com estado *pending* (`d` espera motion) | Vim idêntico em TUI e desktop; implementar duas vezes = refatoração garantida |
| D2 | Buffer: linhas `[][]rune` + gap buffer por linha (estilo micro/kilo) | Simples e correto; interface `Buffer` permite trocar por piece table/rope depois |
| D3 | Posições internas = linha + offset em rune; conversão UTF-16 **só na fronteira do LSP** | Doeu tarde em todo mundo que adiou essa decisão |
| D4 | Highlight MVP via `go/scanner` (stdlib, sem CGO) | Editor Go-only; tree-sitter só se multi-língua virar pauta |
| D5 | Cliente LSP mão própria (~500 linhas do subconjunto) | Entender o protocolo de verdade; biblioteca é avaliação posterior |
| D6 | TUI com **Bubble Tea v2** (Lip Gloss/Bubbles v2) | Stack Go TUI madura e ativa em 2026; precedentes: micro, gocui/lazygit |
| D7 | Desktop com **Wails v3** (estável desde out/2026), shell HTML + área de texto em `<canvas>` único | Já dominado (newpostman); risco do webview isolado no frontend |
| D8 | MVP de viewport: **virtualizar linhas** desde o milestone 3 | Arquivo de 5k linhas é o teste de fogo, não é otimização futura |

## Fase 1 — Terminal (agora)

| MS | Entrega | Critério de aceite |
|---|---|---|
| 0 | Esqueleto do repo + tipos do contrato (`Event`, `Frame`, `StyledLine`, `Style`) | `go build ./...` passa; estrutura existe |
| 1 | **Viewer:** abre arquivo por arg, renderiza, scroll | Navega num `.go` de 5k linhas sem travar |
| 2 | **Buffer + modos:** `i`, `Esc`, digitar, salvar, undo | Edita um `.go` real e o `go build` dele passa |
| 3 | **Highlight** via `Frame/StyledLine` | Código Go colorido no terminal — contrato de render provado |
| 4 | **Vim engine:** motions + operadores (`d/c/y` + text objects simples) | `ciw`, `dap`, `dd`, counts funcionam |
| 5 | **gopls via LSP:** diagnósticos, hover, completion | Erros do compilador em tempo real enquanto digita |
| 6 | **Polimento:** status line, busca, abas | Algo de postar no r/golang |

## Fase 2 — Desktop (depois do MS6)

| MS | Entrega | Critério de aceite |
|---|---|---|
| 7 | Spike descartável: canvas no Wails renderizando `Frame` | Digitação sem lag num arquivo de 5k linhas |
| 8 | Shell Wails v3: janela, abas, file tree, menus | O mesmo core roda no desktop |
| 9 | Integração + polimento | Distribuição binária pra Windows/macOS/Linux |

## Cortes deliberados da v1 (listas pra dizer "não")

Debugger · plugin system · terminal embarcado · git UI completo · colaboração/CRDT · multi-língua · tree-sitter · features LSP além do top-5 (call hierarchy, semantic tokens, inlay hints)

## Gatilhos de reavaliação

- **Mês 3–4 sem nada rodável visível:** recortar milestone, não adiar.
- **Canvas falhar no spike (MS7):** avaliar Gio (gvcode como referência) — o core porta, só o shell muda.
- **Motivação inverter para sistemas distribuídos:** ver `C:\Projetos\WAILS\newpostman\reports\Editor Go ou tarefa distribuida.md` (gatilhos de pivot pro projeto B).

## Acordo de trabalho (guia)

- Eu (assistente) guio: objetivo da sessão, pontos de atenção, critério de aceite.
- Você constrói; cola o código pra review quando quiser — o papel da IA é apontar problemas de design **antes** que virem refatoração.
- Referência de pesquisa completa: `C:\Projetos\WAILS\newpostman\reports\Editor Go ou tarefa distribuida.md` (+ notas em `research_notes/`).
