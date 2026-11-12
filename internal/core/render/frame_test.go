package render

import "testing"

func TestFrameCreationAndDefaults(t *testing.T) {
	var defaultStyle Style
	if defaultStyle.Fg.Kind != ColorDefault || defaultStyle.Bg.Kind != ColorDefault {
		t.Fatalf("esperava ColorDefault no zero value, obteve Fg=%+v Bg=%+v", defaultStyle.Fg, defaultStyle.Bg)
	}
	if defaultStyle.Bold || defaultStyle.Italic || defaultStyle.Underline || defaultStyle.Undercurl {
		t.Fatalf("esperava todas as flags false no zero value, obteve %+v", defaultStyle)
	}

	if ColorRed.Kind != ColorANSI || ColorRed.Index != 1 {
		t.Errorf("esperava ColorRed como ANSI(1), obteve %+v", ColorRed)
	}
	custom := RGB(255, 85, 85)
	if custom.Kind != ColorRGB || custom.R != 255 || custom.G != 85 || custom.B != 85 {
		t.Errorf("RGB incorreto: obteve %+v", custom)
	}

	line := StyledLine{
		{Text: "func", Style: Style{Fg: ColorBlue, Bold: true}},
		{Text: " ", Style: Style{}},
		{Text: "main", Style: Style{Fg: ColorYellow}},
		{Text: "() {", Style: Style{}},
	}

	frame := Frame{
		Lines:  []StyledLine{line},
		Cursor: ScreenPos{Row: 0, Col: 12},
		Status: StyledLine{{Text: "-- INSERT --  main.go"}},
	}

	if len(frame.Lines) != 1 {
		t.Fatalf("esperava 1 linha, obteve %d", len(frame.Lines))
	}
	if len(frame.Lines[0]) != 4 {
		t.Fatalf("esperava 4 spans na linha, obteve %d", len(frame.Lines[0]))
	}
	if frame.Cursor.Row != 0 || frame.Cursor.Col != 12 {
		t.Errorf("cursor incorreto: obteve %+v, esperava {Row:0 Col:12}", frame.Cursor)
	}
	if frame.Lines[0][1].Style != (Style{}) {
		t.Errorf("esperava que o segundo span usasse o estilo padrão")
	}
}
