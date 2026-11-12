package syntax

import (
	"go/scanner"
	"go/token"
	"strings"
	"unicode/utf8"
)

type Kind uint8

const (
	KindPlain Kind = iota
	KindKeyword
	KindString
	KindNumber
	KindComment
	KindPredeclared
	KindFunc
)

type State uint8

const (
	StateNormal State = iota
	StateBlockComment
	StateRawString
)

type Token struct {
	Start int
	End   int
	Kind  Kind
}

var predeclared = map[string]bool{
	"bool": true, "byte": true, "complex64": true, "complex128": true, "error": true,
	"float32": true, "float64": true, "int": true, "int8": true, "int16": true,
	"int32": true, "int64": true, "rune": true, "string": true, "uint": true,
	"uint8": true, "uint16": true, "uint32": true, "uint64": true, "uintptr": true,
	"any": true, "comparable": true,
	"true": true, "false": true, "iota": true, "nil": true,
	"append": true, "cap": true, "clear": true, "close": true, "complex": true,
	"copy": true, "delete": true, "imag": true, "len": true, "make": true,
	"max": true, "min": true, "new": true, "panic": true, "print": true,
	"println": true, "real": true, "recover": true,
}

func HighlightLine(line []rune, in State) ([]Token, State) {
	src := string(line)
	toRune := runeIndex(src)
	var tokens []Token
	offset := 0

	switch in {
	case StateBlockComment:
		end := strings.Index(src, "*/")
		if end < 0 {
			return appendToken(tokens, 0, len(line), KindComment), StateBlockComment
		}
		offset = end + 2
		tokens = appendToken(tokens, 0, toRune[offset], KindComment)
	case StateRawString:
		end := strings.IndexByte(src, '`')
		if end < 0 {
			return appendToken(tokens, 0, len(line), KindString), StateRawString
		}
		offset = end + 1
		tokens = appendToken(tokens, 0, toRune[offset], KindString)
	}

	rest := src[offset:]
	if rest == "" {
		return tokens, StateNormal
	}

	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(rest))
	var s scanner.Scanner
	s.Init(file, []byte(rest), func(token.Position, string) {}, scanner.ScanComments)

	type scanned struct {
		start, end int
		tok        token.Token
		lit        string
	}
	var all []scanned
	out := StateNormal

	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.SEMICOLON && lit == "\n" {
			continue
		}
		start := offset + file.Offset(pos)
		text := lit
		if text == "" {
			text = tok.String()
		}
		end := start + len(text)
		if end > len(src) {
			end = len(src)
		}
		all = append(all, scanned{start: start, end: end, tok: tok, lit: lit})

		switch {
		case tok == token.COMMENT && strings.HasPrefix(lit, "/*") && (len(lit) < 4 || !strings.HasSuffix(lit, "*/")):
			out = StateBlockComment
		case tok == token.STRING && strings.HasPrefix(lit, "`") && (len(lit) < 2 || !strings.HasSuffix(lit, "`")):
			out = StateRawString
		}
	}

	for i, t := range all {
		kind := KindPlain
		switch {
		case t.tok.IsKeyword():
			kind = KindKeyword
		case t.tok == token.STRING || t.tok == token.CHAR:
			kind = KindString
		case t.tok == token.INT || t.tok == token.FLOAT || t.tok == token.IMAG:
			kind = KindNumber
		case t.tok == token.COMMENT:
			kind = KindComment
		case t.tok == token.IDENT && predeclared[t.lit]:
			kind = KindPredeclared
		case t.tok == token.IDENT && i+1 < len(all) && all[i+1].tok == token.LPAREN:
			kind = KindFunc
		}
		if kind != KindPlain {
			tokens = appendToken(tokens, toRune[t.start], toRune[t.end], kind)
		}
	}

	return tokens, out
}

func appendToken(tokens []Token, start, end int, kind Kind) []Token {
	if end <= start {
		return tokens
	}
	return append(tokens, Token{Start: start, End: end, Kind: kind})
}

func runeIndex(s string) []int {
	idx := make([]int, len(s)+1)
	r := 0
	for i := 0; i < len(s); {
		_, size := utf8.DecodeRuneInString(s[i:])
		for j := 0; j < size; j++ {
			idx[i+j] = r
		}
		i += size
		r++
	}
	idx[len(s)] = r
	return idx
}
