package state

import "fmt"

type State uint8

const (
	UNKNOWN State = iota
	NORMAL
	NO_SEMICOLON
	IN_STATEMENT
)

func (s State) String() string {
	switch s {
	case UNKNOWN:
		return "UNKNOWN"
	case NORMAL:
		return "NORMAL"
	case NO_SEMICOLON:
		return "NO_SEMICOLON"
	case IN_STATEMENT:
		return "IN_STATEMENT"
	default:
		panic(fmt.Errorf("unknown state: %d", s))
	}
}
