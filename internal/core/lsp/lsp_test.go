package lsp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
)

func TestReadFrame(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "simples", input: "Content-Length: 2\r\n\r\n{}", want: "{}"},
		{name: "cabecalho extra", input: "Content-Length: 4\r\nContent-Type: x\r\n\r\nnull", want: "null"},
		{name: "nome em minusculas", input: "content-length: 2\r\n\r\n[]", want: "[]"},
		{name: "utf-8 conta bytes", input: "Content-Length: 7\r\n\r\n\"ção\"", want: "\"ção\""},
		{name: "sem content-length", input: "X: 1\r\n\r\n{}", wantErr: true},
		{name: "corpo truncado", input: "Content-Length: 10\r\n\r\n{}", wantErr: true},
		{name: "tamanho invalido", input: "Content-Length: abc\r\n\r\n{}", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := readFrame(bufio.NewReader(strings.NewReader(tt.input)))
			if tt.wantErr {
				if err == nil {
					t.Errorf("esperava erro, veio %q", body)
				}
				return
			}
			if err != nil || string(body) != tt.want {
				t.Errorf("readFrame = %q, %v; esperava %q", body, err, tt.want)
			}
		})
	}
}

func TestUTF16(t *testing.T) {
	line := []rune("a😀ção世x")
	tests := []struct {
		runeCol int
		utf16   int
	}{
		{0, 0}, {1, 1}, {2, 3}, {3, 4}, {5, 6}, {6, 7}, {7, 8},
	}
	for _, tt := range tests {
		if got := ToPosition(line, buffer.Pos{Col: tt.runeCol}).Character; got != tt.utf16 {
			t.Errorf("ToPosition(col %d) = %d, esperava %d", tt.runeCol, got, tt.utf16)
		}
		if got := FromPosition(line, Position{Character: tt.utf16}).Col; got != tt.runeCol {
			t.Errorf("FromPosition(%d) = %d, esperava %d", tt.utf16, got, tt.runeCol)
		}
	}
	if got := FromPosition(line, Position{Character: 999}).Col; got != len(line) {
		t.Errorf("posicao alem do fim deveria parar no fim, veio %d", got)
	}
	if got := FromPosition(line, Position{Character: 2}).Col; got != 2 {
		t.Errorf("posicao no meio do emoji (surrogate) deveria ir para a rune seguinte, veio %d", got)
	}
}

func TestURI(t *testing.T) {
	var path, want string
	if runtime.GOOS == "windows" {
		path, want = `C:\Projetos\meu app\main.go`, "file:///C:/Projetos/meu%20app/main.go"
	} else {
		path, want = "/home/eu/meu app/main.go", "file:///home/eu/meu%20app/main.go"
	}
	uri := PathToURI(path)
	if uri != want {
		t.Errorf("PathToURI = %q, esperava %q", uri, want)
	}
	if got := URIToPath(uri); got != path {
		t.Errorf("URIToPath = %q, esperava %q", got, path)
	}
	if runtime.GOOS == "windows" && !SameURI(uri, "file:///c:/Projetos/meu%20app/main.go") {
		t.Error("no Windows, a letra do drive nao deveria importar")
	}
}

func TestHoverResultFormatos(t *testing.T) {
	tests := map[string]string{
		`{"kind":"markdown","value":"func Println"}`: "func Println",
		`"texto puro"`:                        "texto puro",
		`[{"language":"go","value":"a"},"b"]`: "a\n\nb",
	}
	for raw, want := range tests {
		if got := (hoverResult{Contents: json.RawMessage(raw)}).text(); got != want {
			t.Errorf("hover %s = %q, esperava %q", raw, got, want)
		}
	}
}

func TestCompletionResultFormatos(t *testing.T) {
	var a, b completionResult
	if err := json.Unmarshal([]byte(`[{"label":"x"}]`), &a); err != nil || len(a.Items) != 1 {
		t.Errorf("lista: %+v %v", a, err)
	}
	if err := json.Unmarshal([]byte(`{"isIncomplete":true,"items":[{"label":"y"},{"label":"z"}]}`), &b); err != nil || len(b.Items) != 2 {
		t.Errorf("CompletionList: %+v %v", b, err)
	}
}

type fakeServer struct {
	conn      *Conn
	mu        sync.Mutex
	notes     []string
	texts     []string
	shutdown  bool
	configLen int
}

func (s *fakeServer) record(method string, params json.RawMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notes = append(s.notes, method)
	if method == "textDocument/didOpen" || method == "textDocument/didChange" {
		var p struct {
			TextDocument struct {
				Text string `json:"text"`
			} `json:"textDocument"`
			ContentChanges []struct {
				Text string `json:"text"`
			} `json:"contentChanges"`
		}
		_ = json.Unmarshal(params, &p)
		text := p.TextDocument.Text
		if len(p.ContentChanges) > 0 {
			text = p.ContentChanges[0].Text
		}
		s.texts = append(s.texts, text)
	}
}

func (s *fakeServer) snapshot() ([]string, []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.notes...), append([]string(nil), s.texts...)
}

func startFake(t *testing.T) (*Client, *fakeServer) {
	t.Helper()
	srv := &fakeServer{}
	starter := func(root string) (Transport, error) {
		clientR, serverW := io.Pipe()
		serverR, clientW := io.Pipe()
		srv.conn = NewConn(serverR, serverW, Handler{
			Notification: srv.record,
			Request: func(method string, params json.RawMessage) (any, error) {
				switch method {
				case "initialize":
					go func() {
						var cfg []any
						_ = srv.conn.Call(context.Background(), "workspace/configuration",
							map[string]any{"items": []any{map[string]any{}, map[string]any{}}}, &cfg)
						srv.mu.Lock()
						srv.configLen = len(cfg)
						srv.mu.Unlock()
					}()
					return map[string]any{"capabilities": map[string]any{}}, nil
				case "textDocument/hover":
					return map[string]any{"contents": map[string]any{"kind": "markdown", "value": "```go\nfunc Println()\n```"}}, nil
				case "textDocument/completion":
					return map[string]any{"items": []any{map[string]any{"label": "Println"}, map[string]any{"label": "Printf"}}}, nil
				case "shutdown":
					srv.mu.Lock()
					srv.shutdown = true
					srv.mu.Unlock()
					return nil, nil
				}
				return nil, errors.New("metodo desconhecido")
			},
		})
		return Transport{Reader: clientR, Writer: clientW}, nil
	}
	c := Start("main.go", starter)
	t.Cleanup(c.Close)
	return c, srv
}

func waitEvent[T Event](t *testing.T, c *Client) T {
	t.Helper()
	timeout := time.After(5 * time.Second)
	for {
		select {
		case ev := <-c.Events():
			if v, ok := ev.(T); ok {
				return v
			}
		case <-timeout:
			var zero T
			t.Fatalf("timeout esperando %T", zero)
			return zero
		}
	}
}

func TestClientComServidorFalso(t *testing.T) {
	c, srv := startFake(t)

	c.Open("main.go", "package main\n", 1)
	if st := waitEvent[StatusEvent](t, c); !st.Ready || st.Err != nil {
		t.Fatalf("status inicial = %+v", st)
	}

	c.Change("main.go", "versao 2", 2)
	c.Change("main.go", "versao 3", 3)
	c.Change("main.go", "versao 4", 4)

	c.Hover(7, "main.go", []rune("fmt.Println()"), buffer.Pos{Col: 5})
	h := waitEvent[HoverEvent](t, c)
	if h.ID != 7 || !strings.Contains(h.Text, "func Println()") || h.Err != nil {
		t.Errorf("hover = %+v", h)
	}

	c.Completion(8, "main.go", []rune("fmt.Pr"), buffer.Pos{Col: 6})
	comp := waitEvent[CompletionEvent](t, c)
	if comp.ID != 8 || len(comp.Items) != 2 || comp.Items[0].Label != "Println" {
		t.Errorf("completion = %+v", comp)
	}

	srv.conn.Notify("textDocument/publishDiagnostics", map[string]any{
		"uri":     PathToURI("main.go"),
		"version": 4,
		"diagnostics": []any{map[string]any{
			"range":    map[string]any{"start": map[string]any{"line": 0, "character": 0}, "end": map[string]any{"line": 0, "character": 7}},
			"severity": 1,
			"message":  "erro de teste",
		}},
	})
	srv.conn.Notify("textDocument/publishDiagnostics", map[string]any{
		"uri":         PathToURI("outro.go"),
		"diagnostics": []any{},
	})
	d := waitEvent[DiagnosticsEvent](t, c)
	if !SameURI(d.URI, PathToURI("main.go")) {
		t.Errorf("diagnostico deveria vir marcado com o arquivo: %q", d.URI)
	}
	if len(d.Items) != 1 || d.Items[0].Message != "erro de teste" || d.Items[0].Severity != SeverityError {
		t.Errorf("diagnosticos = %+v", d)
	}

	notes, texts := srv.snapshot()
	if len(notes) < 2 || notes[0] != "initialized" || notes[1] != "textDocument/didOpen" {
		t.Errorf("ordem das notificacoes = %v", notes)
	}
	if texts[0] != "package main\n" || texts[len(texts)-1] != "versao 4" {
		t.Errorf("textos enviados = %q", texts)
	}
	if len(texts) > 3 {
		t.Errorf("mudancas seguidas deveriam ser agrupadas; enviou %d textos: %q", len(texts), texts)
	}

	srv.mu.Lock()
	cfg := srv.configLen
	srv.mu.Unlock()
	if cfg != 2 {
		t.Errorf("resposta a workspace/configuration deveria ter 2 itens, teve %d", cfg)
	}

	c.Close()
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if !srv.shutdown {
		t.Error("Close deveria mandar shutdown ao servidor")
	}
}

func TestClientSemGopls(t *testing.T) {
	c := Start("main.go", func(string) (Transport, error) { return Transport{}, ErrGoplsNotFound })
	defer c.Close()
	st := waitEvent[StatusEvent](t, c)
	if !errors.Is(st.Err, ErrGoplsNotFound) {
		t.Errorf("status = %+v", st)
	}
	c.Change("main.go", "x", 2)
	c.Hover(1, "main.go", nil, buffer.Pos{})
}

func TestConnServidorMorre(t *testing.T) {
	clientR, serverW := io.Pipe()
	conn := NewConn(clientR, io.Discard, Handler{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		serverW.Close()
	}()

	errCh := make(chan error, 1)
	go func() { errCh <- conn.Call(context.Background(), "x", nil, nil) }()

	select {
	case err := <-errCh:
		if !errors.Is(err, ErrClosed) {
			t.Errorf("Call depois do servidor morrer = %v, esperava ErrClosed", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Call ficou travado depois que o servidor morreu")
	}
	if err := conn.Call(context.Background(), "y", nil, nil); !errors.Is(err, ErrClosed) {
		t.Errorf("Call numa conexao fechada = %v", err)
	}
}

func TestWriteFrame(t *testing.T) {
	var out bytes.Buffer
	c := &Conn{w: &out, pending: map[int64]chan response{}, done: make(chan struct{})}
	if err := c.Notify("initialized", map[string]any{}); err != nil {
		t.Fatal(err)
	}
	body, err := readFrame(bufio.NewReader(&out))
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != `{"jsonrpc":"2.0","method":"initialized","params":{}}` {
		t.Errorf("corpo = %s", body)
	}
}
