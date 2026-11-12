package editor

import (
	"strings"

	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

var logoFont = map[rune][5]string{
	'G': {" ████", "█    ", "█  ██", "█   █", " ████"},
	'O': {" ███ ", "█   █", "█   █", "█   █", " ███ "},
	'T': {"█████", "  █  ", "  █  ", "  █  ", "  █  "},
	'E': {"█████", "█    ", "████ ", "█    ", "█████"},
	'X': {"█   █", " █ █ ", "  █  ", " █ █ ", "█   █"},
	'F': {"█████", "█    ", "████ ", "█    ", "█    "},
	'R': {"████ ", "█   █", "████ ", "█  █ ", "█   █"},
}

const logoWord = "GOTEXTFORGE"

var introHints = [][2]string{
	{"i", "começar a escrever"},
	{":q", "sair"},
	{"gotextforge arquivo.go", "abrir um arquivo"},
}

func (e *Editor) showIntro() bool {
	return e.path == "" && e.mode != ModeInsert && !e.buf.Modified() &&
		e.buf.LineCount() == 1 && e.buf.LineLen(0) == 0
}

func logoRows() [5]string {
	var rows [5]string
	for i := range rows {
		parts := make([]string, 0, len(logoWord))
		for _, r := range logoWord {
			parts = append(parts, logoFont[r][i])
		}
		rows[i] = strings.Join(parts, " ")
	}
	return rows
}

func gradient(t float64) render.Color {
	stops := []render.Color{colBlue, colMagenta, colRed}
	if t <= 0 {
		return stops[0]
	}
	if t >= 1 {
		return stops[len(stops)-1]
	}
	seg := t * float64(len(stops)-1)
	i := int(seg)
	f := seg - float64(i)
	a, b := stops[i], stops[i+1]
	mix := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*f) }
	return render.RGB(mix(a.R, b.R), mix(a.G, b.G), mix(a.B, b.B))
}

func (e *Editor) overlayIntro(lines []render.StyledLine) {
	var block []render.StyledLine

	rows := logoRows()
	logoWidth := len([]rune(rows[0]))
	if e.width >= logoWidth+2 {
		for _, row := range rows {
			var line render.StyledLine
			for x, r := range []rune(row) {
				style := render.Style{}
				if r != ' ' {
					style = render.Style{Fg: gradient(float64(x) / float64(logoWidth-1))}
				}
				line = append(line, render.Span{Text: string(r), Style: style})
			}
			block = append(block, line)
		}
		block = append(block, nil)
	} else {
		block = append(block, render.StyledLine{{Text: "gotextforge", Style: render.Style{Fg: colBlue, Bold: true}}}, nil)
	}

	block = append(block,
		render.StyledLine{{Text: "editor modal para Go", Style: styleIntroSub}},
		nil,
	)
	keyWidth, descWidth := 0, 0
	for _, h := range introHints {
		keyWidth = max(keyWidth, len([]rune(h[0])))
		descWidth = max(descWidth, len([]rune(h[1])))
	}
	for _, h := range introHints {
		block = append(block, render.StyledLine{
			{Text: padRight(h[0], keyWidth+3), Style: styleIntroKey},
			{Text: padRight(h[1], descWidth), Style: styleIntroDesc},
		})
	}

	if len(block)+2 > len(lines) {
		return
	}
	top := (len(lines) - len(block)) / 2
	if top < 1 {
		top = 1
	}
	for i, content := range block {
		left := (e.width - content.Width()) / 2
		if left < 0 {
			continue
		}
		row := render.StyledLine{{Text: "~", Style: styleTilde}}
		if pad := left - 1; pad > 0 {
			row = append(row, render.Span{Text: strings.Repeat(" ", pad)})
		}
		lines[top+i] = truncateLine(append(row, content...), e.width)
	}
}

func padRight(s string, n int) string {
	if w := len([]rune(s)); w < n {
		return s + strings.Repeat(" ", n-w)
	}
	return s
}
