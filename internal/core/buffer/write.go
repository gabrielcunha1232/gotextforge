package buffer

import (
	"bufio"
	"io"
	"unicode/utf8"
)

func (b *Buffer) WriteTo(w io.Writer) (int64, error) {
	bw := bufio.NewWriter(w)
	var n int64

	write := func(s string) error {
		m, err := bw.WriteString(s)
		n += int64(m)
		return err
	}

	sep := "\n"
	if b.lineEnding == LineEndingCRLF {
		sep = "\r\n"
	}

	if b.hasBOM {
		if err := write(string(bomUTF8)); err != nil {
			return n, err
		}
	}

	for i, line := range b.lines {
		if i > 0 {
			if err := write(sep); err != nil {
				return n, err
			}
		}
		if err := write(string(line)); err != nil {
			return n, err
		}
	}

	if b.finalNewline {
		if err := write(sep); err != nil {
			return n, err
		}
	}

	return n, bw.Flush()
}

func (b *Buffer) PlainText() string {
	n := len(b.lines)
	for _, l := range b.lines {
		for _, r := range l {
			if uint32(r) < utf8.RuneSelf {
				n++
			} else {
				n += runeLen(r)
			}
		}
	}
	buf := make([]byte, 0, n)
	for i, l := range b.lines {
		if i > 0 {
			buf = append(buf, '\n')
		}
		for _, r := range l {
			if uint32(r) < utf8.RuneSelf {
				buf = append(buf, byte(r))
			} else {
				buf = utf8.AppendRune(buf, r)
			}
		}
	}
	if b.finalNewline {
		buf = append(buf, '\n')
	}
	return string(buf)
}

func (b *Buffer) HasInvalidUTF8() bool {
	return b.invalidUTF8
}

func runeLen(r rune) int {
	if n := utf8.RuneLen(r); n > 0 {
		return n
	}
	return utf8.RuneLen(utf8.RuneError)
}
