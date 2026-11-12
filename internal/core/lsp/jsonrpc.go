package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

var ErrClosed = errors.New("conexão LSP fechada")

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("LSP erro %d: %s", e.Code, e.Message)
}

type incoming struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *RPCError       `json:"error,omitempty"`
}

type outRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type outNotification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type outResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result"`
	Error   *RPCError       `json:"error,omitempty"`
}

type response struct {
	result json.RawMessage
	err    error
}

type Handler struct {
	Notification func(method string, params json.RawMessage)
	Request      func(method string, params json.RawMessage) (any, error)
}

type Conn struct {
	w   io.Writer
	wmu sync.Mutex

	mu      sync.Mutex
	nextID  int64
	pending map[int64]chan response
	closed  bool

	handler Handler
	done    chan struct{}
	readErr error
}

func NewConn(r io.Reader, w io.Writer, h Handler) *Conn {
	c := &Conn{
		w:       w,
		pending: make(map[int64]chan response),
		handler: h,
		done:    make(chan struct{}),
	}
	go c.readLoop(bufio.NewReader(r))
	return c
}

func (c *Conn) Done() <-chan struct{} {
	return c.done
}

func (c *Conn) Call(ctx context.Context, method string, params, result any) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return ErrClosed
	}
	c.nextID++
	id := c.nextID
	ch := make(chan response, 1)
	c.pending[id] = ch
	c.mu.Unlock()

	if err := c.write(outRequest{JSONRPC: "2.0", ID: id, Method: method, Params: params}); err != nil {
		c.forget(id)
		return err
	}

	select {
	case resp := <-ch:
		if resp.err != nil {
			return resp.err
		}
		if result == nil || len(resp.result) == 0 || string(resp.result) == "null" {
			return nil
		}
		return json.Unmarshal(resp.result, result)
	case <-ctx.Done():
		c.forget(id)
		return ctx.Err()
	case <-c.done:
		return ErrClosed
	}
}

func (c *Conn) Notify(method string, params any) error {
	return c.write(outNotification{JSONRPC: "2.0", Method: method, Params: params})
}

func (c *Conn) forget(id int64) {
	c.mu.Lock()
	delete(c.pending, id)
	c.mu.Unlock()
}

func (c *Conn) write(msg any) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	c.wmu.Lock()
	defer c.wmu.Unlock()
	if _, err := fmt.Fprintf(c.w, "Content-Length: %d\r\n\r\n", len(body)); err != nil {
		return err
	}
	_, err = c.w.Write(body)
	return err
}

func (c *Conn) readLoop(r *bufio.Reader) {
	defer c.shutdown()
	for {
		body, err := readFrame(r)
		if err != nil {
			c.readErr = err
			return
		}
		var msg incoming
		if err := json.Unmarshal(body, &msg); err != nil {
			continue
		}
		c.dispatch(msg)
	}
}

func (c *Conn) dispatch(msg incoming) {
	hasID := len(msg.ID) > 0 && string(msg.ID) != "null"
	switch {
	case msg.Method != "" && hasID:
		go c.answer(msg)
	case msg.Method != "":
		if c.handler.Notification != nil {
			c.handler.Notification(msg.Method, msg.Params)
		}
	case hasID:
		id, err := strconv.ParseInt(string(msg.ID), 10, 64)
		if err != nil {
			return
		}
		c.mu.Lock()
		ch := c.pending[id]
		delete(c.pending, id)
		c.mu.Unlock()
		if ch == nil {
			return
		}
		if msg.Error != nil {
			ch <- response{err: msg.Error}
		} else {
			ch <- response{result: msg.Result}
		}
	}
}

func (c *Conn) answer(msg incoming) {
	var result any
	var err error
	if c.handler.Request != nil {
		result, err = c.handler.Request(msg.Method, msg.Params)
	}
	resp := outResponse{JSONRPC: "2.0", ID: msg.ID, Result: result}
	if err != nil {
		resp.Result = nil
		resp.Error = &RPCError{Code: -32601, Message: err.Error()}
	}
	_ = c.write(resp)
}

func (c *Conn) shutdown() {
	c.mu.Lock()
	c.closed = true
	pending := c.pending
	c.pending = map[int64]chan response{}
	c.mu.Unlock()
	for _, ch := range pending {
		ch <- response{err: ErrClosed}
	}
	close(c.done)
}

func readFrame(r *bufio.Reader) ([]byte, error) {
	length := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		name, value, ok := strings.Cut(line, ":")
		if ok && strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			n, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil || n < 0 {
				return nil, fmt.Errorf("Content-Length inválido: %q", value)
			}
			length = n
		}
	}
	if length < 0 {
		return nil, errors.New("mensagem LSP sem Content-Length")
	}
	if length > 64<<20 {
		return nil, fmt.Errorf("mensagem LSP grande demais: %d bytes", length)
	}
	body := make([]byte, length)
	_, err := io.ReadFull(r, body)
	return body, err
}
