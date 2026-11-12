# gotextforge

Editor de texto modal para Go, no terminal, no estilo do Neovim/LazyVim. É escrito em Go puro sobre um núcleo sem nenhuma dependência de interface, e foi pensado para ganhar depois um frontend desktop (Wails) que reaproveita o mesmo núcleo.

```
            ████  ███  █████ █████ █   █ █████ █████  ███  ████   ████ █████
           █     █   █   █   █      █ █    █   █     █   █ █   █ █     █
           █  ██ █   █   █   ████    █     █   ████  █   █ ████  █  ██ ████
           █   █ █   █   █   █      █ █    █   █     █   █ █  █  █   █ █
            ████  ███    █   █████ █   █   █   █      ███  █   █  ████ █████
```

## O que ele faz

- **Edição vim:** motions (`w b e f t { } gg G`...), operadores `d c y` com qualquer motion, text objects (`ciw`, `da"`, `di(`, `yap`...), counts (`3dd`, `2d3w`), `p P J r x`, e undo/redo agrupados por comando.
- **gopls integrado:**
  - erros do compilador **enquanto você digita** (sublinhado ondulado e a mensagem no fim da linha);
  - `K` para ver a documentação numa janela flutuante;
  - completion automático ao digitar.
- **Syntax highlight** de Go via `go/scanner`. Comentários de bloco e raw strings que atravessam linhas são tratados corretamente: a validação foi feita contra toda a stdlib do Go (≈3 milhões de linhas, zero divergências).
- **Vários arquivos em abas**, uma busca `/` com destaque enquanto você digita, e uma status line no estilo lualine.
- **Rápido em arquivos grandes:** só as linhas visíveis são desenhadas. Um quadro leva menos de 1 ms, e abrir um arquivo de 110 mil linhas não trava.
- **Seguro com os seus arquivos:**
  - gravação atômica (o arquivo temporário só substitui o original no fim);
  - preserva CRLF, BOM e o newline final do arquivo;
  - recusa salvar um arquivo com UTF-8 inválido em vez de corrompê-lo;
  - nunca envia ao terminal os caracteres de controle que estão dentro do arquivo.

## Instalar e rodar

Requer Go 1.27+ (veja o `go.mod`) e, para os recursos de LSP, o [gopls](https://pkg.go.dev/golang.org/x/tools/gopls):

```
go install golang.org/x/tools/gopls@latest
go run ./cmd/gotextforge                  # tela inicial
go run ./cmd/gotextforge main.go util.go  # abre arquivos em abas
```

O tema usa truecolor e foi feito para terminais escuros (como o terminal do VS Code e o Windows Terminal).

## Atalhos

| Modo normal | |
|---|---|
| `h j k l`, `w b e`, `0 ^ $`, `gg G`, `f t F T`, `{ }` | movimento (aceitam count: `5j`) |
| `d c y` + movimento ou text object | apagar, trocar, copiar (`dw`, `ci"`, `yap`) |
| `dd cc yy`, `x X D C s S Y` | atalhos de linha e caractere |
| `p P`, `J`, `r<c>` | colar, juntar linhas, substituir caractere |
| `i a I A o O` | modo insert |
| `u`, `Ctrl+r` | desfazer, refazer |
| `/texto`, `?texto`, `n N`, `* #` | buscar (`:noh` apaga o destaque) |
| `K` | documentação (gopls) |
| `H L`, `gt gT` | aba anterior e próxima |

| Modo insert | |
|---|---|
| `Esc` | voltar ao modo normal |
| `Ctrl+n` / `Ctrl+p`, setas | navegar no completion |
| `Enter` / `Tab` / `Ctrl+y` | aceitar o completion |
| `Ctrl+Espaço` | pedir o completion manualmente |

| Comandos | |
|---|---|
| `:w`, `:wa` | salvar, salvar todos |
| `:e arquivo` | abrir (ou criar) em uma aba nova |
| `:q`, `:q!` | fechar a aba (sai na última) |
| `:wq`, `:x` | salvar e fechar |
| `:qa`, `:qa!` | sair de tudo |
| `:bn`, `:bp` | próxima e anterior aba |

## Arquitetura

```
teclado → frontend (Bubble Tea v2) → input.Event → CORE → render.Frame → frontend → tela
```

O núcleo (`internal/core/`) não importa nada de UI. O frontend só traduz teclas e desenha quadros (`Frame`: linhas com spans estilizados, cursor e status line). Cada pacote cuida de uma coisa:

| Pacote | Função |
|---|---|
| `buffer` | o texto, a edição e o undo |
| `view` | o layout da tela |
| `syntax` | o highlight |
| `engine` | a gramática vim, as motions e os text objects |
| `lsp` | o cliente JSON-RPC próprio para o gopls |
| `storage` | a gravação atômica |
| `editor` | junta tudo |

Mais detalhes das decisões em [`ROADMAP.md`](ROADMAP.md) e [`HANDOFF.md`](HANDOFF.md).

## Testes

```
go test -short ./...   # rápido, sem gopls
go test ./...          # inclui os testes de integração com o gopls real
```
