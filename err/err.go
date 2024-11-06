package err

import (
	"fmt"
	"strings"

	h "github.com/fholmqvist/remlisp/highlight"
	tk "github.com/fholmqvist/remlisp/token"
)

type Error struct {
	Msg   string
	Start int
	End   int
}

func (e Error) String(input []byte) string {
	// TODO: Take input and calculate row/col.
	if e.Start == e.End {
		e.End += 1
	}
	rowi := 0
	row := h.Bold(fmt.Sprintf(" %d |", rowi))
	var errline strings.Builder
	errline.WriteString(h.Code(string(input[:e.Start])))
	errline.WriteString(h.ErrorCode(string(input[e.Start:e.End])))
	errline.WriteString(h.Code(string(input[e.End:])))
	return fmt.Sprintf("%s %s\n\n%s", row, errline.String(), e.Msg)
}

func FromToken(t tk.Token, msg string) *Error {
	pos := t.Pos()
	return &Error{
		Msg:   msg,
		Start: pos.Start,
		End:   pos.End,
	}
}

func FromPosition(p tk.Position, msg string) *Error {
	return &Error{
		Msg:   msg,
		Start: p.Start,
		End:   p.End,
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
