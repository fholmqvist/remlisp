package repl

import (
	"fmt"

	"slices"

	h "github.com/fholmqvist/remlisp/highlight"
)

type line struct {
	line    []byte
	history [][]byte
	cursor  int
	hidx    int8
}

func newLine() *line {
	return &line{
		line:    []byte{},
		history: [][]byte{},
		cursor:  0,
		hidx:    0,
	}
}

func (l *line) get() []byte {
	if len(l.line) > 0 {
		l.history = append(l.history, l.line)
		if len(l.history) > MAX_HISTORY {
			l.history = l.history[1:]
		}
	}
	return l.line
}

func (l *line) reset() {
	l.line = []byte{}
	l.cursor = 0
	l.hidx = 0
}

func (l *line) add(b []byte) {
	l.line = append(l.line[:l.cursor], append(b, l.line[l.cursor:]...)...)
	l.cursor += len(b)
}

func (l *line) openParen() {
	l.add([]byte("()"))
	l.left()
}

func (l *line) openBracket() {
	l.add([]byte("[]"))
	l.left()
}

func (l *line) openBrace() {
	l.add([]byte("{}"))
	l.left()
}

func (l *line) quotation() {
	l.add([]byte(`""`))
	l.left()
}

func (l *line) left() {
	if l.cursor > 0 {
		l.cursor--
	}
}

func (l *line) right() {
	if l.cursor < len(l.line) {
		l.cursor++
	}
}

func (l *line) leftWord() {
	for i := l.cursor - 1; i >= 0; i-- {
		if l.line[i] == ' ' {
			l.cursor = i + 1
			return
		}
	}
	l.cursor = 0
}

func (l *line) rightWord() {
	for i := l.cursor; i < len(l.line); i++ {
		if l.line[i] == ' ' {
			l.cursor = i
			return
		}
	}
	l.cursor = len(l.line)
}

func (l *line) home() {
	l.cursor = 0
}

func (l *line) end() {
	l.cursor = len(l.line)
}

func (l *line) backspace() {
	if l.cursor > 0 {
		curr := l.line[l.cursor-1]
		hasNext := l.cursor < len(l.line)
		if hasNext {
			next := l.line[l.cursor]
			if curr == '(' && next == ')' {
				l.line = slices.Delete(l.line, l.cursor-1, l.cursor+1)
			} else if curr == '[' && next == ']' {
				l.line = slices.Delete(l.line, l.cursor-1, l.cursor+1)
			} else if curr == '{' && next == '}' {
				l.line = slices.Delete(l.line, l.cursor-1, l.cursor+1)
			} else if curr == '"' && next == '"' {
				l.line = slices.Delete(l.line, l.cursor-1, l.cursor+1)
			} else {
				l.line = slices.Delete(l.line, l.cursor-1, l.cursor)
			}
		} else {
			l.line = slices.Delete(l.line, l.cursor-1, l.cursor)
		}
		l.cursor--
	}
}

func (l *line) delete() {
	if l.cursor == len(l.line) {
		return
	}
	l.line = slices.Delete(l.line, l.cursor, l.cursor+1)
}

func (l *line) backspaceWord() {
	if l.cursor <= 0 {
		return
	}
	for i := l.cursor - 1; i >= 0; i-- {
		if l.line[i] == ' ' {
			l.line = append(l.line[:i], l.line[l.cursor:]...)
			l.cursor = i
			return
		}
	}
}

func (l *line) deleteWord() {
	if l.cursor >= len(l.line) {
		return
	}
	for i := l.cursor; i < len(l.line); i++ {
		if l.line[i] == ' ' {
			l.line = append(l.line[:l.cursor], l.line[i:]...)
			return
		}
	}
}

func (l *line) prevHistory() {
	l.hidx--
	l.useHistory()
}

func (l *line) nextHistory() {
	l.hidx++
	l.useHistory()
}

func (l *line) useHistory() {
	if len(l.history) == 0 {
		l.reset()
		l.hidx = 0
		return
	}
	if l.hidx < 0 {
		l.hidx = int8(len(l.history) - 1)
	}
	if l.hidx > int8(len(l.history))-1 {
		l.hidx = 0
	}
	l.line = []byte(l.history[l.hidx])
	l.end()
}

func (l *line) print() {
	fmt.Print("\r\033[K")
	fmt.Printf("> %v", h.Code(string(l.line)))
	diff := max(0, len(l.line)-l.cursor)
	for i := 0; i < diff; i++ {
		fmt.Print("\033[D")
	}
}
