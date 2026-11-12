package buffer

import (
	"bufio"
	"io"
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

func (b *Buffer) HasInvalidUTF8() bool {
	return b.invalidUTF8
}
