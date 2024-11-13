package repl

import (
	"fmt"
	"os"
)

const (
	UP             = `\x1b[A`
	DOWN           = `\x1b[B`
	LEFT           = `\x1b[D`
	RIGHT          = `\x1b[C`
	CTRL_LEFT      = `\x1b[1;5D`
	CTRL_RIGHT     = `\x1b[1;5C`
	SHIFT_LEFT     = `\x1b[1;2D`
	SHIFT_RIGHT    = `\x1b[1;2C`
	HOME           = `\x1b[H`
	END            = `\x1b[F`
	CTRL_HOME      = `\x1b[1;5H`
	CTRL_END       = `\x1b[1;5F`
	BACKSPACE      = `\x7f`
	DELETE         = `\x1b[3~`
	CTRL_BACKSPACE = `\x17`
	CTRL_DELETE    = `\x1bd`
	OPEN_PAREN     = `(`
	OPEN_BRACKET   = `[`
	OPEN_BRACE     = `{`
	QUOTATION      = `\"`
)

func (r *Repl) input() []byte {
	var buf [12]byte
	r.line.reset()
	r.line.print()
OUTER:
	for {
		n, err := os.Stdin.Read(buf[:])
		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}
		cmd := fmt.Sprintf("%q", buf[:n])
		cmd = cmd[1 : len(cmd)-1]
		switch cmd {
		case UP:
			r.line.prevHistory()
		case DOWN:
			r.line.nextHistory()
		case LEFT:
			r.line.left()
		case RIGHT:
			r.line.right()
		case CTRL_LEFT:
			r.line.leftWord()
		case CTRL_RIGHT:
			r.line.rightWord()
		case SHIFT_LEFT:
			// TODO: Fake selection.
		case SHIFT_RIGHT:
			// TODO: Fake selection.
		case HOME, CTRL_HOME:
			r.line.home()
		case END, CTRL_END:
			r.line.end()
		case BACKSPACE:
			r.line.backspace()
		case DELETE:
			r.line.delete()
		case CTRL_BACKSPACE:
			r.line.backspaceWord()
		case CTRL_DELETE:
			r.line.deleteWord()
		case OPEN_PAREN:
			r.line.openParen()
		case OPEN_BRACKET:
			r.line.openBracket()
		case OPEN_BRACE:
			r.line.openBrace()
		case QUOTATION:
			r.line.quotation()
		case `\n`, `\r`:
			break OUTER
		default:
			r.line.add(buf[:n])
		}
		r.line.print()
	}
	return r.line.get()
}
