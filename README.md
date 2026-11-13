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
- **Navegação pelo projeto sem digitar caminhos:**
  - explorador lateral (como o neo-tree) para criar, renomear e apagar arquivos e pastas;
  - busca fuzzy de arquivos (como o Telescope): `edtgo` acha `editor/editor.go`;
  - tela inicial com ações de uma tecla.
- **Painel de comandos:**
  - roda `go build`, `go test`, `docker compose up` e outros sem sair do editor, com a saída aparecendo enquanto o comando roda;
  - **Enter num erro (`arquivo.go:12:5`) abre o arquivo na linha**;
  - Ctrl+c cancela.
- **Edição confortável:** fecha parênteses, colchetes, chaves e aspas sozinho, e indenta ao dar Enter (`{|}` + Enter abre o bloco já indentado).
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
go run ./cmd/gotextforge                  # tela inicial na pasta atual
go run ./cmd/gotextforge ~/meu-projeto    # abre a pasta com o explorador
go run ./cmd/gotextforge main.go util.go  # abre arquivos em abas
```

O tema usa truecolor e foi feito para terminais escuros (como o terminal do VS Code e o Windows Terminal).

## Navegando pelo projeto

A tecla líder é o **Espaço**, como no LazyVim.

| Atalho | |
|---|---|
| `Espaço Espaço` ou `Espaço f f` | buscar arquivo pelo nome (fuzzy); se não existir, oferece **+ criar** |
| `Espaço e` | abrir, focar ou fechar o explorador |
| `Espaço f n` | novo arquivo (já sugere a pasta do arquivo atual) |
| `Espaço ,` | lista as abas abertas |
| `Espaço b d` | fechar a aba |

| No explorador | |
|---|---|
| `j k`, `g G` | mover |
| `l` / `Enter` | entrar na pasta ou abrir o arquivo |
| `h` | fechar a pasta ou ir para a pasta pai |
| `a` | criar arquivo (termine com `/` para criar uma pasta; `a/b/c.go` cria tudo) |
| `r`, `d` | renomear, apagar (com confirmação; pasta só se estiver vazia) |
| `Backspace` | subir uma pasta (navega pelo disco inteiro) |
| `.` | usar a pasta selecionada como pasta do projeto |
| `H`, `R` | mostrar ou ocultar arquivos ocultos, atualizar |
| `Esc`, `q` | voltar ao editor, fechar o explorador |

Na tela inicial: `f` busca, `e` explora, `n` cria arquivo, `p` troca de pasta, `q` sai.

## Rodando comandos

| Atalho | Comando |
|---|---|
| `:!comando` ou `Espaço t n` | roda qualquer comando na pasta do projeto |
| `Espaço g b` / `g t` / `g r` / `g v` / `g m` | `go build ./...` / `go test ./...` / `go run .` / `go vet ./...` / `go mod tidy` |
| `Espaço d u` / `d d` / `d p` / `d l` | `docker compose up` / `down` / `docker ps` / `logs -f` |
| `Espaço t t` | focar, abrir ou fechar o painel |
| `Espaço t r` | repetir o último comando |

| No painel | |
|---|---|
| `j k`, `g G` | percorrer a saída |
| `n` / `N` | próximo / anterior erro com arquivo:linha |
| `Enter` | abrir o arquivo do erro na linha certa |
| `Ctrl+c` | cancelar o comando |
| `r`, `i` | repetir, digitar um novo comando |
| `Esc`, `q` | voltar ao editor, fechar o painel |

O painel não é um terminal completo: ele **não recebe teclado**. Programas interativos (`docker exec -it`, um `go run` que lê da entrada) precisam do seu terminal.

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
