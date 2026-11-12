package engine

import (
	"strings"

	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
)

type waitKind uint8

const (
	waitNone waitKind = iota
	waitChar
	waitG
	waitObject
)

type Parser struct {
	count1  int
	count2  int
	op      Operator
	opKey   rune
	wait    waitKind
	waitFor MotionKind
	waitAct Action
	inner   bool
	keys    []string
}

const countLimit = 99999

func (p *Parser) Reset() {
	*p = Parser{}
}

func (p *Parser) Pending() string {
	return strings.Join(p.keys, "")
}

func (p *Parser) Feed(k input.Key) (Command, bool) {
	if k.Code == input.KeyEsc {
		p.Reset()
		return Command{}, false
	}
	if k.Code == input.KeyRune && k.Mod == input.ModCtrl {
		pending := len(p.keys) > 0
		p.Reset()
		switch {
		case k.Rune == 'c' && !pending:
			return Command{Action: ActionQuit}, true
		case k.Rune == 'r' && !pending:
			return Command{Action: ActionRedo}, true
		}
		return Command{}, false
	}
	if k.Mod != 0 {
		p.Reset()
		return Command{}, false
	}

	if k.Code != input.KeyRune {
		return p.feedSpecial(k)
	}
	r := k.Rune
	p.keys = append(p.keys, string(r))

	switch p.wait {
	case waitChar:
		if p.waitAct == ActionReplace {
			return p.finish(Command{Action: ActionReplace, Char: r})
		}
		return p.finishMotion(Motion{Kind: p.waitFor, Char: r})
	case waitG:
		switch {
		case r == 'g':
			return p.finishMotion(Motion{Kind: MotionFileStart})
		case r == 't' && p.op == OpNone:
			return p.finish(Command{Action: ActionNextTab})
		case r == 'T' && p.op == OpNone:
			return p.finish(Command{Action: ActionPrevTab})
		}
		return p.cancel()
	case waitObject:
		obj, ok := objectForRune(r)
		if !ok {
			return p.cancel()
		}
		return p.finishMotion(Motion{Kind: MotionTextObject, Object: obj, Inner: p.inner})
	}

	if r >= '1' && r <= '9' || r == '0' && p.currentCount() > 0 {
		p.addDigit(int(r - '0'))
		return Command{}, false
	}

	if p.op != OpNone {
		switch {
		case r == p.opKey:
			return p.finishMotion(Motion{Kind: MotionCurrentLine})
		case r == 'i' || r == 'a':
			p.wait = waitObject
			p.inner = r == 'i'
			return Command{}, false
		}
		if m, ok := p.motionForRune(r); ok {
			return p.finishMotion(m)
		}
		if p.wait != waitNone {
			return Command{}, false
		}
		return p.cancel()
	}

	switch r {
	case 'd':
		return p.startOperator(OpDelete, r)
	case 'c':
		return p.startOperator(OpChange, r)
	case 'y':
		return p.startOperator(OpYank, r)
	case 'x':
		return p.finishShorthand(OpDelete, Motion{Kind: MotionRight})
	case 'X':
		return p.finishShorthand(OpDelete, Motion{Kind: MotionLeft})
	case 'D':
		return p.finishShorthand(OpDelete, Motion{Kind: MotionLineEnd})
	case 'C':
		return p.finishShorthand(OpChange, Motion{Kind: MotionLineEnd})
	case 's':
		return p.finishShorthand(OpChange, Motion{Kind: MotionRight})
	case 'S':
		return p.finishShorthand(OpChange, Motion{Kind: MotionCurrentLine})
	case 'Y':
		return p.finishShorthand(OpYank, Motion{Kind: MotionCurrentLine})
	case 'i':
		return p.finish(Command{Action: ActionInsert})
	case 'a':
		return p.finish(Command{Action: ActionAppend})
	case 'I':
		return p.finish(Command{Action: ActionInsertLineStart})
	case 'A':
		return p.finish(Command{Action: ActionAppendLineEnd})
	case 'o':
		return p.finish(Command{Action: ActionOpenBelow})
	case 'O':
		return p.finish(Command{Action: ActionOpenAbove})
	case 'p':
		return p.finish(Command{Action: ActionPasteAfter})
	case 'P':
		return p.finish(Command{Action: ActionPasteBefore})
	case 'u':
		return p.finish(Command{Action: ActionUndo})
	case 'J':
		return p.finish(Command{Action: ActionJoin})
	case ':':
		return p.finish(Command{Action: ActionCommandLine})
	case 'K':
		return p.finish(Command{Action: ActionHover})
	case '/':
		return p.finish(Command{Action: ActionSearchForward})
	case '?':
		return p.finish(Command{Action: ActionSearchBackward})
	case 'n':
		return p.finish(Command{Action: ActionSearchNext})
	case 'N':
		return p.finish(Command{Action: ActionSearchPrev})
	case '*':
		return p.finish(Command{Action: ActionSearchWordForward})
	case '#':
		return p.finish(Command{Action: ActionSearchWordBackward})
	case 'L':
		return p.finish(Command{Action: ActionNextTab})
	case 'H':
		return p.finish(Command{Action: ActionPrevTab})
	case 'r':
		p.wait = waitChar
		p.waitAct = ActionReplace
		return Command{}, false
	}

	if m, ok := p.motionForRune(r); ok {
		return p.finishMotion(m)
	}
	if p.wait != waitNone {
		return Command{}, false
	}
	return p.cancel()
}

func (p *Parser) feedSpecial(k input.Key) (Command, bool) {
	if p.wait != waitNone {
		return p.cancel()
	}
	var m MotionKind
	switch k.Code {
	case input.KeyLeft, input.KeyBackspace:
		m = MotionLeft
	case input.KeyRight:
		m = MotionRight
	case input.KeyDown, input.KeyEnter:
		m = MotionDown
	case input.KeyUp:
		m = MotionUp
	case input.KeyHome:
		m = MotionLineStart
	case input.KeyEnd:
		m = MotionLineEnd
	case input.KeyPageDown:
		m = MotionPageDown
	case input.KeyPageUp:
		m = MotionPageUp
	case input.KeyDelete:
		if p.op != OpNone {
			return p.cancel()
		}
		return p.finishShorthand(OpDelete, Motion{Kind: MotionRight})
	default:
		return p.cancel()
	}
	return p.finishMotion(Motion{Kind: m})
}

func (p *Parser) motionForRune(r rune) (Motion, bool) {
	switch r {
	case 'h':
		return Motion{Kind: MotionLeft}, true
	case 'l', ' ':
		return Motion{Kind: MotionRight}, true
	case 'j':
		return Motion{Kind: MotionDown}, true
	case 'k':
		return Motion{Kind: MotionUp}, true
	case 'w':
		return Motion{Kind: MotionWordStart}, true
	case 'W':
		return Motion{Kind: MotionBigWordStart}, true
	case 'e':
		return Motion{Kind: MotionWordEnd}, true
	case 'E':
		return Motion{Kind: MotionBigWordEnd}, true
	case 'b':
		return Motion{Kind: MotionWordBack}, true
	case 'B':
		return Motion{Kind: MotionBigWordBack}, true
	case '0':
		return Motion{Kind: MotionLineStart}, true
	case '^':
		return Motion{Kind: MotionFirstNonBlank}, true
	case '$':
		return Motion{Kind: MotionLineEnd}, true
	case 'G':
		return Motion{Kind: MotionFileEnd}, true
	case '}':
		return Motion{Kind: MotionParagraphForward}, true
	case '{':
		return Motion{Kind: MotionParagraphBackward}, true
	case 'g':
		p.wait = waitG
	case 'f':
		p.wait, p.waitFor = waitChar, MotionFindForward
	case 't':
		p.wait, p.waitFor = waitChar, MotionTillForward
	case 'F':
		p.wait, p.waitFor = waitChar, MotionFindBackward
	case 'T':
		p.wait, p.waitFor = waitChar, MotionTillBackward
	}
	return Motion{}, false
}

func objectForRune(r rune) (ObjectKind, bool) {
	switch r {
	case 'w':
		return ObjectWord, true
	case 'W':
		return ObjectBigWord, true
	case 'p':
		return ObjectParagraph, true
	case '"':
		return ObjectDoubleQuote, true
	case '\'':
		return ObjectSingleQuote, true
	case '`':
		return ObjectBacktick, true
	case '(', ')', 'b':
		return ObjectParen, true
	case '[', ']':
		return ObjectBracket, true
	case '{', '}', 'B':
		return ObjectBrace, true
	case '<', '>':
		return ObjectAngle, true
	}
	return ObjectNone, false
}

func (p *Parser) currentCount() int {
	if p.op == OpNone {
		return p.count1
	}
	return p.count2
}

func (p *Parser) addDigit(d int) {
	c := &p.count1
	if p.op != OpNone {
		c = &p.count2
	}
	if *c*10+d <= countLimit {
		*c = *c*10 + d
	}
}

func (p *Parser) startOperator(op Operator, key rune) (Command, bool) {
	p.op = op
	p.opKey = key
	return Command{}, false
}

func (p *Parser) totalCount() int {
	switch {
	case p.count1 == 0:
		return p.count2
	case p.count2 == 0:
		return p.count1
	default:
		return min(p.count1*p.count2, countLimit)
	}
}

func (p *Parser) finishMotion(m Motion) (Command, bool) {
	return p.finish(Command{Operator: p.op, Motion: m})
}

func (p *Parser) finishShorthand(op Operator, m Motion) (Command, bool) {
	return p.finish(Command{Operator: op, Motion: m})
}

func (p *Parser) finish(c Command) (Command, bool) {
	c.Count = p.totalCount()
	p.Reset()
	return c, true
}

func (p *Parser) cancel() (Command, bool) {
	p.Reset()
	return Command{}, false
}
