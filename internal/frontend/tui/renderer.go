package tui

import (
	"strconv"
	"strings"

	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

const sgrReset = "\x1b[0m"

func RenderFrame(frame render.Frame) string {
	var b strings.Builder
	b.Grow(len(frame.Lines) * 128)

	for _, line := range frame.Lines {
		for _, span := range line {
			writeSpan(&b, span.Text, span.Style)
		}
		b.WriteByte('\n')
	}

	for _, span := range frame.Status {
		writeSpan(&b, span.Text, span.Style)
	}

	return b.String()
}

func writeSpan(b *strings.Builder, text string, s render.Style) {
	if s == (render.Style{}) {
		b.WriteString(text)
		return
	}
	b.WriteString("\x1b[")
	first := true
	code := func(c string) {
		if !first {
			b.WriteByte(';')
		}
		b.WriteString(c)
		first = false
	}

	if s.Bold {
		code("1")
	}
	if s.Italic {
		code("3")
	}
	if s.Underline || s.Undercurl {
		code("4")
	}
	if c := colorCode(s.Fg, false); c != "" {
		code(c)
	}
	if c := colorCode(s.Bg, true); c != "" {
		code(c)
	}

	b.WriteByte('m')
	b.WriteString(text)
	b.WriteString(sgrReset)
}

func colorCode(c render.Color, background bool) string {
	switch c.Kind {
	case render.ColorANSI:
		i := int(c.Index)
		switch {
		case i < 8 && background:
			return strconv.Itoa(40 + i)
		case i < 8:
			return strconv.Itoa(30 + i)
		case i < 16 && background:
			return strconv.Itoa(100 + i - 8)
		case i < 16:
			return strconv.Itoa(90 + i - 8)
		case background:
			return "48;5;" + strconv.Itoa(i)
		default:
			return "38;5;" + strconv.Itoa(i)
		}
	case render.ColorRGB:
		prefix := "38;2;"
		if background {
			prefix = "48;2;"
		}
		return prefix + strconv.Itoa(int(c.R)) + ";" + strconv.Itoa(int(c.G)) + ";" + strconv.Itoa(int(c.B))
	default:
		return ""
	}
}
