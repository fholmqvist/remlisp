package transpiler

import (
	"fmt"
	"strings"

	"github.com/fholmqvist/remlisp/transpiler/state"
)

var DEBUG_STATE = false

func (t *Transpiler) setState(s state.State) {
	t.oldstate = append(t.oldstate, t.state)
	if DEBUG_STATE {
		fmt.Printf("move: %s -> %s | %v\n", t.state, s, t.oldstate)
	}
	t.state = s
}

func (t *Transpiler) hasState(s state.State) bool {
	for _, s2 := range t.oldstate {
		if s2 == s {
			return true
		}
	}
	return false
}

func (t *Transpiler) restoreState() {
	old := t.oldstate[len(t.oldstate)-1]
	t.oldstate = t.oldstate[:len(t.oldstate)-1]
	if DEBUG_STATE {
		fmt.Printf("back: %s -> %s | %v\n", t.state, old, t.oldstate)
	}
	t.state = old
}

func fixName(s string) string {
	s = strings.ReplaceAll(s, "->", "_arrow_")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, "?", "P")
	s = strings.ReplaceAll(s, "!", "Ex")
	return s
}
