package compiler

import (
	"fmt"
	"strings"

	"github.com/fholmqvist/remlisp/compiler/state"
)

var DEBUG_STATE = false

func (c *Compiler) setState(s state.State) {
	c.oldstate = append(c.oldstate, c.state)
	if DEBUG_STATE {
		fmt.Printf("move: %s -> %s | %v\n", c.state, s, c.oldstate)
	}
	c.state = s
}

func (c *Compiler) restoreState() {
	old := c.oldstate[len(c.oldstate)-1]
	c.oldstate = c.oldstate[:len(c.oldstate)-1]
	if DEBUG_STATE {
		fmt.Printf("back: %s -> %s | %v\n", c.state, old, c.oldstate)
	}
	c.state = old
}

func fixName(s string) string {
	s = strings.ReplaceAll(s, "->", "_arrow_")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, "?", "P")
	s = strings.ReplaceAll(s, "!", "Ex")
	return s
}
