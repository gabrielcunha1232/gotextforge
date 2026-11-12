package buffer

type Pos struct {
	Line int
	Col  int
}

func (p Pos) Before(q Pos) bool {
	if p.Line != q.Line {
		return p.Line < q.Line
	}
	return p.Col < q.Col
}
