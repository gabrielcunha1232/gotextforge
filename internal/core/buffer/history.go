package buffer

type changeKind uint8

const (
	changeInsert changeKind = iota
	changeDelete
)

type change struct {
	kind  changeKind
	start Pos
	end   Pos
	text  string
}

type group struct {
	id      int
	cursor  Pos
	changes []change
}

type history struct {
	undo      []*group
	redo      []*group
	current   *group
	open      bool
	openAt    Pos
	nextID    int
	savedID   int
	replaying bool
	dirty     bool
	dirtyFrom int
}

func (b *Buffer) TakeDirty() (int, bool) {
	if !b.dirty {
		return 0, false
	}
	b.dirty = false
	return b.dirtyFrom, true
}

func (b *Buffer) BeginChange(cursor Pos) {
	if b.open {
		return
	}
	b.open = true
	b.openAt = cursor
	b.current = nil
}

func (b *Buffer) EndChange() {
	b.open = false
	b.current = nil
}

func (b *Buffer) record(c change) {
	if !b.dirty || c.start.Line < b.dirtyFrom {
		b.dirty = true
		b.dirtyFrom = c.start.Line
	}
	if b.replaying {
		return
	}
	if b.current == nil {
		b.nextID++
		cursor := c.start
		if b.open {
			cursor = b.openAt
		}
		b.current = &group{id: b.nextID, cursor: cursor}
		b.undo = append(b.undo, b.current)
		b.redo = b.redo[:0]
	}
	b.current.changes = append(b.current.changes, c)
	if !b.open {
		b.current = nil
	}
}

func (b *Buffer) Undo() (Pos, bool) {
	b.EndChange()
	if len(b.undo) == 0 {
		return Pos{}, false
	}
	g := b.undo[len(b.undo)-1]
	b.undo = b.undo[:len(b.undo)-1]

	b.replaying = true
	for i := len(g.changes) - 1; i >= 0; i-- {
		c := g.changes[i]
		switch c.kind {
		case changeInsert:
			b.Delete(c.start, c.end)
		case changeDelete:
			b.Insert(c.start, c.text)
		}
	}
	b.replaying = false

	b.redo = append(b.redo, g)
	return b.ClampPos(g.cursor), true
}

func (b *Buffer) Redo() (Pos, bool) {
	b.EndChange()
	if len(b.redo) == 0 {
		return Pos{}, false
	}
	g := b.redo[len(b.redo)-1]
	b.redo = b.redo[:len(b.redo)-1]

	b.replaying = true
	for _, c := range g.changes {
		switch c.kind {
		case changeInsert:
			b.Insert(c.start, c.text)
		case changeDelete:
			b.Delete(c.start, c.end)
		}
	}
	b.replaying = false

	b.undo = append(b.undo, g)
	return b.ClampPos(g.changes[0].start), true
}

func (b *Buffer) Modified() bool {
	return b.topID() != b.savedID
}

func (b *Buffer) MarkSaved() {
	b.savedID = b.topID()
}

func (b *Buffer) topID() int {
	if len(b.undo) == 0 {
		return 0
	}
	return b.undo[len(b.undo)-1].id
}
