package state

import "fmt"

type State uint8

const (
	UNKNOWN State = iota
	NORMAL
	THREADING
)

func (s State) String() string {
	switch s {
	case UNKNOWN:
		return "UNKNOWN"
	case NORMAL:
		return "NORMAL"
	case THREADING:
		return "THREADING"
	default:
		panic(fmt.Errorf("unknown state: %d", s))
	}
}
