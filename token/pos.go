package token

import "fmt"

type Position struct {
	Start int
	End   int
}

func NewPos(start, end int) Position {
	if start == end {
		end++
	}
	return Position{
		Start: start,
		End:   end,
	}
}

func (p Position) String() string {
	// TODO: Take input and calculate row/col.
	return fmt.Sprintf("[byte index %d-%d]", p.Start+1, p.End+1)
}

func Between(a, b Position) Position {
	return Position{
		Start: a.Start,
		End:   b.End,
	}
}
