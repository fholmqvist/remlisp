package state

type State uint8

const (
	UNKNOWN State = iota
	NORMAL
	NO_SEMICOLON
)
