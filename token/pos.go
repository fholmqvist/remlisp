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
	return fmt.Sprintf("[byte index %d-%d]", p.Start+1, p.End+1)
}

// Creates a new token that starts from a and ends at b.
func Between(a, b Position) Position {
	return Position{
		Start: a.Start,
		End:   b.End,
	}
}
