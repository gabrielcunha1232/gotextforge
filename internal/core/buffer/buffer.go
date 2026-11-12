package buffer

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

var ErrUTF16 = errors.New("arquivo em UTF-16 não é suportado")

var (
	bomUTF8    = []byte{0xEF, 0xBB, 0xBF}
	bomUTF16LE = []byte{0xFF, 0xFE}
	bomUTF16BE = []byte{0xFE, 0xFF}
)

type LineEnding uint8

const (
	LineEndingLF LineEnding = iota
	LineEndingCRLF
)

type Buffer struct {
	lines      [][]rune
	lineEnding LineEnding
	hasBOM     bool
}

func New() *Buffer {
	return &Buffer{
		lines:      [][]rune{{}},
		lineEnding: LineEndingLF,
	}
}

func Load(r io.Reader) (*Buffer, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if bytes.HasPrefix(data, bomUTF16LE) || bytes.HasPrefix(data, bomUTF16BE) {
		return nil, ErrUTF16
	}

	b := &Buffer{
		lineEnding: LineEndingLF,
	}

	if bytes.HasPrefix(data, bomUTF8) {
		b.hasBOM = true
		data = data[len(bomUTF8):]
	}

	if len(data) == 0 {
		b.lines = [][]rune{{}}
		return b, nil
	}

	if bytes.Contains(data, []byte("\r\n")) {
		b.lineEnding = LineEndingCRLF
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)

	for scanner.Scan() {
		b.lines = append(b.lines, []rune(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(b.lines) == 0 {
		b.lines = [][]rune{{}}
	}

	return b, nil
}

func (b *Buffer) LineCount() int {
	return len(b.lines)
}

func (b *Buffer) Line(i int) []rune {
	if i < 0 || i >= len(b.lines) {
		return nil
	}
	out := make([]rune, len(b.lines[i]))
	copy(out, b.lines[i])
	return out
}

func (b *Buffer) LineLen(i int) int {
	if i < 0 || i >= len(b.lines) {
		return 0
	}
	return len(b.lines[i])
}

func (b *Buffer) LineEnding() LineEnding {
	return b.lineEnding
}

func (b *Buffer) HasBOM() bool {
	return b.hasBOM
}
