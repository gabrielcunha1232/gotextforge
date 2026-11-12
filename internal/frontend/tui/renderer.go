package tui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

func RenderFrame(frame render.Frame, height int) string {
	var b strings.Builder

	for i, line := range frame.Lines {
		if i > 0 {
			b.WriteByte('\n')
		}
		for _, span := range line {
			b.WriteString(applyStyle(span.Style).Render(span.Text))
		}
	}

	contentLines := len(frame.Lines)
	if contentLines == 0 {
		contentLines = 1
	}
	if height > contentLines+1 {
		b.WriteString(strings.Repeat("\n", height-contentLines-1))
	} else {
		b.WriteByte('\n')
	}

	statusStyle := lipgloss.NewStyle().Reverse(true)
	b.WriteString(statusStyle.Render(frame.Status))

	return b.String()
}

func applyStyle(s render.Style) lipgloss.Style {
	ls := lipgloss.NewStyle()

	if c, ok := toLipglossColor(s.Fg); ok {
		ls = ls.Foreground(c)
	}
	if c, ok := toLipglossColor(s.Bg); ok {
		ls = ls.Background(c)
	}
	if s.Bold {
		ls = ls.Bold(true)
	}
	if s.Italic {
		ls = ls.Italic(true)
	}

	if s.Underline || s.Undercurl {
		ls = ls.Underline(true)
	}

	return ls
}

func toLipglossColor(c render.Color) (color.Color, bool) {
	switch c.Kind {
	case render.ColorANSI:
		return lipgloss.Color(strconv.Itoa(int(c.Index))), true
	case render.ColorRGB:
		return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)), true
	default:
		return nil, false
	}
}
