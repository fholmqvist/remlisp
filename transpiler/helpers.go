package transpiler

import (
	"fmt"
	"strings"

	"github.com/fholmqvist/remlisp/transpiler/state"
)

var DEBUG_STATE = false

func (t *Transpiler) setState(s state.State) {
	t.state = append(t.state, s)
	if DEBUG_STATE {
		fmt.Printf("move: %s -> %s | %v\n", t.state, s, s)
	}
}

func (t *Transpiler) hasState(s state.State) bool {
	for _, s2 := range t.state {
		if s2 == s {
			return true
		}
	}
	return false
}

func (t *Transpiler) restoreState() {
	old := t.state[len(t.state)-1]
	t.state = t.state[:len(t.state)-1]
	if DEBUG_STATE {
		fmt.Printf("back: %s -> %s | %v\n", t.state, old, t.state)
	}
}

func fixName(s string) string {
	s = strings.ReplaceAll(s, "->>", "_darrow_")
	s = strings.ReplaceAll(s, "->", "_arrow_")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, "?", "P")
	s = strings.ReplaceAll(s, "!", "Ex")
	return s
}
