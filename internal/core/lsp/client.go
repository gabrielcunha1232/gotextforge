package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
)

type Event interface{ isEvent() }

type StatusEvent struct {
	Ready bool
	Err   error
}

type DiagnosticsEvent struct {
	URI     string
	Version int
	Items   []Diagnostic
}

type HoverEvent struct {
	ID   int
	Text string
	Err  error
}

type CompletionEvent struct {
	ID    int
	Items []CompletionItem
	Err   error
}

func (StatusEvent) isEvent()      {}
func (DiagnosticsEvent) isEvent() {}
func (HoverEvent) isEvent()       {}
func (CompletionEvent) isEvent()  {}

type Transport struct {
	Reader io.ReadCloser
	Writer io.WriteCloser
	Wait   func() error
	Kill   func()
}

type Starter func(root string) (Transport, error)

type pendingChange struct {
	text    string
	version int
	queued  bool
}

type Client struct {
	root   string
	events chan Event
	done   chan struct{}
	ended  chan struct{}
	once   sync.Once

	mu             sync.Mutex
	ops            []func(*Conn)
	signal         chan struct{}
	pending        map[string]*pendingChange
	lastSent       map[string]int
	requestTimeout time.Duration
}

var ErrGoplsNotFound = errors.New("gopls não encontrado no PATH (instale com: go install golang.org/x/tools/gopls@latest)")

func GoplsStarter(root string) (Transport, error) {
	bin, err := exec.LookPath("gopls")
	if err != nil {
		return Transport{}, ErrGoplsNotFound
	}
	cmd := exec.Command(bin)
	cmd.Dir = root
	cmd.Stderr = io.Discard
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return Transport{}, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Transport{}, err
	}
	if err := cmd.Start(); err != nil {
		return Transport{}, err
	}
	return Transport{
		Reader: stdout,
		Writer: stdin,
		Wait:   cmd.Wait,
		Kill:   func() { _ = cmd.Process.Kill() },
	}, nil
}

func FindRoot(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Dir(path)
	}
	dir := filepath.Dir(abs)
	for d := dir; ; {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
			return d
		}
		parent := filepath.Dir(d)
		if parent == d {
			return dir
		}
		d = parent
	}
}

func Start(path string, starter Starter) *Client {
	root := FindRoot(path)
	c := &Client{
		root:           root,
		pending:        map[string]*pendingChange{},
		lastSent:       map[string]int{},
		events:         make(chan Event, 64),
		done:           make(chan struct{}),
		ended:          make(chan struct{}),
		signal:         make(chan struct{}, 1),
		requestTimeout: 10 * time.Second,
	}
	go c.run(starter)
	return c
}

func (c *Client) Events() <-chan Event {
	return c.events
}

func (c *Client) Open(path, text string, version int) {
	uri := PathToURI(path)
	c.enqueue(func(conn *Conn) {
		_ = conn.Notify("textDocument/didOpen", map[string]any{
			"textDocument": map[string]any{
				"uri":        uri,
				"languageId": "go",
				"version":    version,
				"text":       text,
			},
		})
		c.lastSent[uri] = version
	})
}

func (c *Client) Change(path, text string, version int) {
	uri := PathToURI(path)
	c.mu.Lock()
	p := c.pending[uri]
	if p == nil {
		p = &pendingChange{}
		c.pending[uri] = p
	}
	p.text, p.version = text, version
	queued := p.queued
	p.queued = true
	c.mu.Unlock()
	if queued {
		return
	}
	c.enqueue(func(conn *Conn) {
		c.mu.Lock()
		text, version := p.text, p.version
		p.queued = false
		c.mu.Unlock()
		if version <= c.lastSent[uri] {
			return
		}
		_ = conn.Notify("textDocument/didChange", map[string]any{
			"textDocument":   map[string]any{"uri": uri, "version": version},
			"contentChanges": []map[string]any{{"text": text}},
		})
		c.lastSent[uri] = version
	})
}

func (c *Client) Save(path string) {
	uri := PathToURI(path)
	c.enqueue(func(conn *Conn) {
		_ = conn.Notify("textDocument/didSave", map[string]any{
			"textDocument": map[string]any{"uri": uri},
		})
	})
}

func (c *Client) CloseDocument(path string) {
	uri := PathToURI(path)
	c.enqueue(func(conn *Conn) {
		_ = conn.Notify("textDocument/didClose", map[string]any{
			"textDocument": map[string]any{"uri": uri},
		})
		delete(c.lastSent, uri)
	})
}

func (c *Client) Hover(id int, path string, line []rune, pos buffer.Pos) {
	params := positionParams(path, line, pos)
	c.enqueue(func(conn *Conn) {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), c.requestTimeout)
			defer cancel()
			var res *hoverResult
			err := conn.Call(ctx, "textDocument/hover", params, &res)
			ev := HoverEvent{ID: id, Err: err}
			if res != nil {
				ev.Text = res.text()
			}
			c.emit(ev)
		}()
	})
}

func (c *Client) Completion(id int, path string, line []rune, pos buffer.Pos) {
	params := positionParams(path, line, pos)
	c.enqueue(func(conn *Conn) {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), c.requestTimeout)
			defer cancel()
			var res completionResult
			err := conn.Call(ctx, "textDocument/completion", params, &res)
			c.emit(CompletionEvent{ID: id, Items: res.Items, Err: err})
		}()
	})
}

func positionParams(path string, line []rune, pos buffer.Pos) map[string]any {
	return map[string]any{
		"textDocument": map[string]any{"uri": PathToURI(path)},
		"position":     ToPosition(line, pos),
	}
}

func (c *Client) Close() {
	c.once.Do(func() { close(c.done) })
	select {
	case <-c.ended:
	case <-time.After(5 * time.Second):
	}
}

func (c *Client) enqueue(op func(*Conn)) {
	c.mu.Lock()
	c.ops = append(c.ops, op)
	c.mu.Unlock()
	select {
	case c.signal <- struct{}{}:
	default:
	}
}

func (c *Client) takeOps() []func(*Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ops := c.ops
	c.ops = nil
	return ops
}

func (c *Client) emit(ev Event) {
	select {
	case c.events <- ev:
	case <-c.done:
	}
}

func (c *Client) run(starter Starter) {
	defer close(c.ended)

	t, err := starter(c.root)
	if err != nil {
		c.emit(StatusEvent{Err: err})
		<-c.done
		return
	}

	conn := NewConn(t.Reader, t.Writer, Handler{
		Notification: c.onNotification,
		Request:      onServerRequest,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	err = conn.Call(ctx, "initialize", c.initializeParams(), nil)
	cancel()
	if err != nil {
		c.emit(StatusEvent{Err: err})
		c.stop(conn, t)
		return
	}
	_ = conn.Notify("initialized", map[string]any{})
	c.emit(StatusEvent{Ready: true})

	for {
		for _, op := range c.takeOps() {
			op(conn)
		}
		select {
		case <-c.signal:
		case <-c.done:
			c.stop(conn, t)
			return
		case <-conn.Done():
			c.emit(StatusEvent{Err: errors.New("gopls encerrou inesperadamente")})
			<-c.done
			return
		}
	}
}

func (c *Client) stop(conn *Conn, t Transport) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	_ = conn.Call(ctx, "shutdown", nil, nil)
	cancel()
	_ = conn.Notify("exit", nil)
	_ = t.Writer.Close()

	waited := make(chan struct{})
	go func() {
		if t.Wait != nil {
			_ = t.Wait()
		}
		close(waited)
	}()
	select {
	case <-waited:
	case <-time.After(2 * time.Second):
		if t.Kill != nil {
			t.Kill()
		}
	}
}

func (c *Client) onNotification(method string, params json.RawMessage) {
	if method != "textDocument/publishDiagnostics" {
		return
	}
	var p struct {
		URI         string       `json:"uri"`
		Version     int          `json:"version"`
		Diagnostics []Diagnostic `json:"diagnostics"`
	}
	if json.Unmarshal(params, &p) != nil {
		return
	}
	c.emit(DiagnosticsEvent{URI: p.URI, Version: p.Version, Items: p.Diagnostics})
}

func onServerRequest(method string, params json.RawMessage) (any, error) {
	if method == "workspace/configuration" {
		var p struct {
			Items []json.RawMessage `json:"items"`
		}
		_ = json.Unmarshal(params, &p)
		return make([]any, len(p.Items)), nil
	}
	return nil, nil
}

func (c *Client) initializeParams() map[string]any {
	rootURI := PathToURI(c.root)
	return map[string]any{
		"processId":  os.Getpid(),
		"clientInfo": map[string]any{"name": "gotextforge"},
		"rootUri":    rootURI,
		"workspaceFolders": []map[string]any{
			{"uri": rootURI, "name": filepath.Base(c.root)},
		},
		"capabilities": map[string]any{
			"general": map[string]any{"positionEncodings": []string{"utf-16"}},
			"workspace": map[string]any{
				"configuration":    true,
				"workspaceFolders": true,
			},
			"textDocument": map[string]any{
				"synchronization":    map[string]any{"didSave": true},
				"hover":              map[string]any{"contentFormat": []string{"markdown", "plaintext"}},
				"completion":         map[string]any{"completionItem": map[string]any{"snippetSupport": false}},
				"publishDiagnostics": map[string]any{},
			},
		},
	}
}
