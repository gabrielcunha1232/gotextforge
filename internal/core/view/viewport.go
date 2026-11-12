package view

type Viewport struct {
	Top    int
	Height int
}

func NewViewport(height int) Viewport {
	if height < 1 {
		height = 1
	}
	return Viewport{
		Top:    0,
		Height: height,
	}
}

func (v *Viewport) Scroll(delta, totalLines int) {
	v.Top += delta
	v.clamp(totalLines)
}

func (v *Viewport) ScrollTop() {
	v.Top = 0
}

func (v *Viewport) ScrollBottom(totalLines int) {
	v.Top = v.maxTop(totalLines)
}

func (v *Viewport) EnsureVisible(line, totalLines int) {
	if line < v.Top {
		v.Top = line
	} else if line >= v.Top+v.Height {
		v.Top = line - v.Height + 1
	}
	v.clamp(totalLines)
}

func (v *Viewport) Resize(height, totalLines int) {
	if height < 1 {
		height = 1
	}
	v.Height = height
	v.clamp(totalLines)
}

func (v Viewport) VisibleRange(totalLines int) (start, end int) {
	if totalLines <= 0 {
		return 0, 0
	}
	start = v.Top
	if start < 0 {
		start = 0
	}
	if start > totalLines {
		start = totalLines
	}

	end = start + v.Height
	if end > totalLines {
		end = totalLines
	}
	return start, end
}

func (v *Viewport) clamp(totalLines int) {
	if v.Top < 0 {
		v.Top = 0
		return
	}
	max := v.maxTop(totalLines)
	if v.Top > max {
		v.Top = max
	}
}

func (v Viewport) maxTop(totalLines int) int {
	if totalLines <= v.Height {
		return 0
	}
	return totalLines - v.Height
}
