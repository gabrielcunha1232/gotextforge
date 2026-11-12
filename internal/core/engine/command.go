package engine

type Operator uint8

const (
	OpNone Operator = iota
	OpDelete
	OpChange
	OpYank
)

type MotionKind uint8

const (
	MotionNone MotionKind = iota
	MotionLeft
	MotionRight
	MotionDown
	MotionUp
	MotionWordStart
	MotionBigWordStart
	MotionWordEnd
	MotionBigWordEnd
	MotionWordBack
	MotionBigWordBack
	MotionLineStart
	MotionFirstNonBlank
	MotionLineEnd
	MotionFileStart
	MotionFileEnd
	MotionFindForward
	MotionTillForward
	MotionFindBackward
	MotionTillBackward
	MotionParagraphForward
	MotionParagraphBackward
	MotionPageDown
	MotionPageUp
	MotionCurrentLine
	MotionTextObject
)

type ObjectKind uint8

const (
	ObjectNone ObjectKind = iota
	ObjectWord
	ObjectBigWord
	ObjectParagraph
	ObjectDoubleQuote
	ObjectSingleQuote
	ObjectBacktick
	ObjectParen
	ObjectBracket
	ObjectBrace
	ObjectAngle
)

type Motion struct {
	Kind   MotionKind
	Char   rune
	Object ObjectKind
	Inner  bool
}

type Action uint8

const (
	ActionNone Action = iota
	ActionInsert
	ActionAppend
	ActionInsertLineStart
	ActionAppendLineEnd
	ActionOpenBelow
	ActionOpenAbove
	ActionPasteAfter
	ActionPasteBefore
	ActionUndo
	ActionRedo
	ActionJoin
	ActionReplace
	ActionCommandLine
	ActionQuit
)

type Command struct {
	Count    int
	Operator Operator
	Motion   Motion
	Action   Action
	Char     rune
}

func (c Command) Times() int {
	if c.Count < 1 {
		return 1
	}
	return c.Count
}
