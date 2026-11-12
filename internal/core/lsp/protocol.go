package lsp

import (
	"encoding/json"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf16"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
)

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Severity int

const (
	SeverityError Severity = iota + 1
	SeverityWarning
	SeverityInfo
	SeverityHint
)

type Diagnostic struct {
	Range    Range    `json:"range"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	Source   string   `json:"source,omitempty"`
}

type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

type CompletionItem struct {
	Label      string    `json:"label"`
	Kind       int       `json:"kind,omitempty"`
	Detail     string    `json:"detail,omitempty"`
	InsertText string    `json:"insertText,omitempty"`
	FilterText string    `json:"filterText,omitempty"`
	SortText   string    `json:"sortText,omitempty"`
	TextEdit   *TextEdit `json:"textEdit,omitempty"`
}

func (c CompletionItem) Text() string {
	switch {
	case c.TextEdit != nil && c.TextEdit.NewText != "":
		return c.TextEdit.NewText
	case c.InsertText != "":
		return c.InsertText
	default:
		return c.Label
	}
}

func (c CompletionItem) Filter() string {
	if c.FilterText != "" {
		return c.FilterText
	}
	return c.Label
}

type completionResult struct {
	Items []CompletionItem
}

func (r *completionResult) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '[' {
		return json.Unmarshal(data, &r.Items)
	}
	var list struct {
		Items []CompletionItem `json:"items"`
	}
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	r.Items = list.Items
	return nil
}

type hoverResult struct {
	Contents json.RawMessage `json:"contents"`
}

func (h hoverResult) text() string {
	var markup struct {
		Value string `json:"value"`
	}
	if json.Unmarshal(h.Contents, &markup) == nil && markup.Value != "" {
		return markup.Value
	}
	var plain string
	if json.Unmarshal(h.Contents, &plain) == nil {
		return plain
	}
	var list []json.RawMessage
	if json.Unmarshal(h.Contents, &list) == nil {
		parts := make([]string, 0, len(list))
		for _, item := range list {
			parts = append(parts, hoverResult{Contents: item}.text())
		}
		return strings.Join(parts, "\n\n")
	}
	return ""
}

func utf16Len(r rune) int {
	if n := utf16.RuneLen(r); n > 0 {
		return n
	}
	return 1
}

func ToPosition(line []rune, p buffer.Pos) Position {
	col := 0
	for i, r := range line {
		if i >= p.Col {
			break
		}
		col += utf16Len(r)
	}
	return Position{Line: p.Line, Character: col}
}

func FromPosition(line []rune, p Position) buffer.Pos {
	units := 0
	for i, r := range line {
		if units >= p.Character {
			return buffer.Pos{Line: p.Line, Col: i}
		}
		units += utf16Len(r)
	}
	return buffer.Pos{Line: p.Line, Col: len(line)}
}

func PathToURI(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	p := filepath.ToSlash(abs)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return (&url.URL{Scheme: "file", Path: p}).String()
}

func URIToPath(uri string) string {
	u, err := url.Parse(uri)
	if err != nil || u.Scheme != "file" {
		return ""
	}
	p := u.Path
	if runtime.GOOS == "windows" && len(p) >= 3 && p[0] == '/' && p[2] == ':' {
		p = p[1:]
	}
	return filepath.FromSlash(p)
}

func SameURI(a, b string) bool {
	pa, pb := URIToPath(a), URIToPath(b)
	if pa == "" || pb == "" {
		return a == b
	}
	if runtime.GOOS == "windows" {
		return strings.EqualFold(pa, pb)
	}
	return pa == pb
}
