package err

import (
	"bytes"
	"fmt"
	"strings"

	h "github.com/fholmqvist/remlisp/highlight"
	tk "github.com/fholmqvist/remlisp/token"
)

type Error struct {
	Msg      string
	RowStart int
	RowEnd   int
	Start    int
	End      int
}

func (e Error) String(input []byte) string {
	row := h.Bold(fmt.Sprintf(" %d |", e.RowStart))
	lines := bytes.Split(input, []byte("\n"))
	line := lines[max(0, e.RowStart-1)]
	if len(line) == 0 && e.RowStart > 1 {
		line = lines[e.RowStart-2]
	}
	// if e.RowStart != e.RowEnd && e.RowStart+1 != e.RowEnd {
	// 	// TODO: This.
	// }
	if e.Start > e.End {
		e.Start, e.End = e.End-1, e.Start
	}
	e.Start = max(0, e.Start-2)
	e.End = min(len(line), max(0, e.End-2))
	var errline strings.Builder
	errline.WriteString(h.Code(string(line[:e.Start])))
	errline.WriteString(h.ErrorCode(string(line[e.Start:e.End])))
	errline.WriteString(h.Code(string(line[e.End:])))
	return fmt.Sprintf("%s %s\n\n%s", row, errline.String(), e.Msg)
}

func FromToken(t tk.Token, msg string) *Error {
	pos := t.Pos()
	return &Error{
		Msg:      msg,
		RowStart: pos.RowStart,
		RowEnd:   pos.RowEnd,
		Start:    pos.Start,
		End:      pos.End,
	}
}

func FromPosition(p tk.Position, msg string) *Error {
	return &Error{
		Msg:      msg,
		RowStart: p.RowStart,
		RowEnd:   p.RowEnd,
		Start:    p.Start,
		End:      p.End,
	}
}

func Panic(reason, msg string) {
	panic(fmt.Sprintf("%s: %s",
		h.Bold(h.Red(reason)), msg))
}

func NotImplemented(reason, msg string) {
	panic(fmt.Sprintf("%s: %s: %s",
		h.Bold(h.Red("not implemented")),
		h.Bold(reason), msg))
}
