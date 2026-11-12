package render

type ColorKind uint8

const (
	ColorDefault ColorKind = iota
	ColorANSI
	ColorRGB
)

var (
	ColorBlack   = ANSI(0)
	ColorRed     = ANSI(1)
	ColorGreen   = ANSI(2)
	ColorYellow  = ANSI(3)
	ColorBlue    = ANSI(4)
	ColorMagenta = ANSI(5)
	ColorCyan    = ANSI(6)
	ColorWhite   = ANSI(7)

	ColorBrightBlack   = ANSI(8)
	ColorBrightRed     = ANSI(9)
	ColorBrightGreen   = ANSI(10)
	ColorBrightYellow  = ANSI(11)
	ColorBrightBlue    = ANSI(12)
	ColorBrightMagenta = ANSI(13)
	ColorBrightCyan    = ANSI(14)
	ColorBrightWhite   = ANSI(15)
)

type Color struct {
	Kind    ColorKind
	Index   uint8
	R, G, B uint8
}

type Style struct {
	Fg        Color
	Bg        Color
	Bold      bool
	Italic    bool
	Underline bool
	Undercurl bool
	Sp        Color
}

type Span struct {
	Text  string
	Style Style
}

type StyledLine []Span

func (l StyledLine) Text() string {
	n := 0
	for _, s := range l {
		n += len(s.Text)
	}
	b := make([]byte, 0, n)
	for _, s := range l {
		b = append(b, s.Text...)
	}
	return string(b)
}

func (l StyledLine) Width() int {
	w := 0
	for _, s := range l {
		for range s.Text {
			w++
		}
	}
	return w
}

type ScreenPos struct {
	Row int
	Col int // em celulas de tela, nao runes: '世' ocupa 2 celulas (wcwidth)
}

type CursorShape uint8

const (
	CursorBlock CursorShape = iota
	CursorBar
)

type Frame struct {
	Lines       []StyledLine
	Cursor      ScreenPos
	CursorShape CursorShape
	HideCursor  bool
	Status      StyledLine
}

// ANSI cria uma cor da paleta de 256 cores: 0–15 básicas, 16–255 estendidas.
func ANSI(index uint8) Color {
	return Color{Kind: ColorANSI, Index: index}
}

func RGB(r, g, b uint8) Color {
	return Color{Kind: ColorRGB, R: r, G: g, B: b}
}
