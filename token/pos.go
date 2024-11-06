package token

import "fmt"

type Position struct {
	RowStart int
	RowEnd   int
	Start    int
	End      int
}

func (p Position) String() string {
	return fmt.Sprintf("[%d:%d-%d]", p.RowStart, p.Start, p.End)
}

func Between(a, b Position) Position {
	return Position{
		RowStart: a.RowStart,
		RowEnd:   b.RowEnd,
		Start:    a.Start,
		End:      b.End,
	}
}
