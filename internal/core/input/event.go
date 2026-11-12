package input

type EventKind uint8

const (
	EventNone EventKind = iota
	EventKey
	EventResize
)

type Mod uint8

const (
	ModCtrl Mod = 1 << iota
	ModAlt
	ModShift
)

type KeyCode uint8

const (
	KeyRune KeyCode = iota
	KeyEsc
	KeyEnter
	KeyBackspace
	KeyTab
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown
	KeyDelete
)

type Key struct {
	Code KeyCode
	Rune rune
	Mod  Mod
}

func (k Key) String() string {
	var base string

	if (k.Code) == KeyRune {
		if k.Rune == 0 {
			return ""
		}
		if k.Rune == ' ' {
			base = "Space"
		} else {
			base = string(k.Rune)
		}
	} else {
		switch k.Code {
		case KeyEsc:
			base = "Esc"
		case KeyEnter:
			base = "CR"
		case KeyBackspace:
			base = "BS"
		case KeyTab:
			base = "Tab"
		case KeyUp:
			base = "Up"
		case KeyDown:
			base = "Down"
		case KeyLeft:
			base = "Left"
		case KeyRight:
			base = "Right"
		case KeyHome:
			base = "Home"
		case KeyEnd:
			base = "End"
		case KeyPageUp:
			base = "PageUp"
		case KeyPageDown:
			base = "PageDown"
		case KeyDelete:
			base = "Del"
		default:
			base = "Unknown"
		}
	}

	var prefix string
	if k.Mod&ModCtrl != 0 {
		prefix += "C-"
	}
	if k.Mod&ModAlt != 0 {
		prefix += "M-"
	}

	if k.Mod&ModShift != 0 && k.Code != KeyRune {
		prefix += "S-"
	}

	if prefix != "" || k.Code != KeyRune || k.Rune == ' ' {
		return "<" + prefix + base + ">"
	}

	return base
}

type Event struct {
	Kind   EventKind
	Key    Key
	Width  int
	Height int
}

func NewKeyEvent(k Key) Event {
	return Event{
		Kind: EventKey,
		Key:  k,
	}
}

func NewResizeEvent(width, height int) Event {
	return Event{
		Kind:   EventResize,
		Width:  width,
		Height: height,
	}
}
