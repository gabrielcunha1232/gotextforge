package editor

import (
	"fmt"
	"strings"

	"github.com/gabrielcunha1232/gotextforge/internal/core/lsp"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

func severityColor(s lsp.Severity) render.Color {
	switch s {
	case lsp.SeverityError:
		return colRed
	case lsp.SeverityWarning:
		return colYellow
	case lsp.SeverityInfo:
		return colBlue1
	default:
		return colTeal
	}
}

type lineDiag struct {
	start, end int
	severity   lsp.Severity
	message    string
}

func (e *Editor) diagnosticsByLine(first, last int) map[int][]lineDiag {
	if len(e.lsp.diags) == 0 {
		return nil
	}
	out := map[int][]lineDiag{}
	for _, d := range e.lsp.diags {
		sl, el := d.Range.Start.Line, d.Range.End.Line
		if el < first || sl > last || sl >= e.buf.LineCount() {
			continue
		}
		el = min(el, e.buf.LineCount()-1)
		for l := max(sl, first); l <= min(el, last); l++ {
			runes := e.buf.Line(l)
			start, end := 0, len(runes)
			if l == sl {
				start = lsp.FromPosition(runes, d.Range.Start).Col
			}
			if l == el {
				end = lsp.FromPosition(runes, d.Range.End).Col
			}
			if end <= start {
				end = min(start+1, len(runes))
				start = min(start, max(len(runes)-1, 0))
			}
			msg := ""
			if l == sl {
				msg = d.Message
			}
			out[l] = append(out[l], lineDiag{start: start, end: end, severity: d.Severity, message: msg})
		}
	}
	return out
}

func worstSeverity(diags []lineDiag) lsp.Severity {
	worst := lsp.Severity(99)
	for _, d := range diags {
		if d.severity > 0 && d.severity < worst {
			worst = d.severity
		}
	}
	if worst == 99 {
		return lsp.SeverityError
	}
	return worst
}

func applyDiagnostics(styles []render.Style, diags []lineDiag) {
	for _, d := range diags {
		for j := max(d.start, 0); j < min(d.end, len(styles)); j++ {
			styles[j].Undercurl = true
			styles[j].Sp = severityColor(d.severity)
		}
	}
}

func virtualText(diags []lineDiag) (string, lsp.Severity) {
	worst := worstSeverity(diags)
	for _, d := range diags {
		if d.severity == worst && d.message != "" {
			msg, _, _ := strings.Cut(d.message, "\n")
			return "  ● " + sanitize(msg), worst
		}
	}
	return "", worst
}

func (e *Editor) diagnosticCounts() (errors, warnings int) {
	for _, d := range e.lsp.diags {
		switch d.Severity {
		case lsp.SeverityError:
			errors++
		case lsp.SeverityWarning:
			warnings++
		}
	}
	return errors, warnings
}

func (e *Editor) lspStatusSpans() render.StyledLine {
	var out render.StyledLine
	errs, warns := e.diagnosticCounts()
	if errs > 0 {
		out = append(out, render.Span{Text: fmt.Sprintf(" E%d", errs), Style: render.Style{Fg: colRed, Bg: colBgStatus, Bold: true}})
	}
	if warns > 0 {
		out = append(out, render.Span{Text: fmt.Sprintf(" W%d", warns), Style: render.Style{Fg: colYellow, Bg: colBgStatus, Bold: true}})
	}

	switch {
	case e.lsp.client == nil:
		if e.highlight {
			out = append(out, render.Span{Text: " go ", Style: styleStatusC})
		}
	case e.lsp.failed:
		out = append(out, render.Span{Text: " gopls! ", Style: styleStatusError})
	case !e.lsp.ready:
		out = append(out, render.Span{Text: " gopls… ", Style: styleStatusC})
	default:
		out = append(out, render.Span{Text: " gopls ", Style: render.Style{Fg: colGreen, Bg: colBgStatus}})
	}
	return out
}
