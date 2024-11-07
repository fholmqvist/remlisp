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
	var (
		row    = 1
		start  strings.Builder
		middle strings.Builder
		end    strings.Builder
	)
	for i := 0; i < len(input); i++ {
		b := input[i]
		var (
			startInner  strings.Builder
			middleInner strings.Builder
			endInner    strings.Builder
		)
		for i < e.Start {
			startInner.WriteByte(b)
			i++
			b = input[i]
			if b == '\n' {
				start.WriteString(h.Bold(fmt.Sprintf("\n %d | ", row)))
				start.WriteString(h.Code(startInner.String()))
				startInner.Reset()
				row++
			}
		}
		// Last row of error is always next
		// line whether it exists or not.
		//
		// 1 | ...
		// 2 | ...
		//
		startRow := row
		for i < e.End {
			middleInner.WriteByte(b)
			i++
			b = input[i]
			if i >= e.End {
				middle.WriteString(h.Bold(fmt.Sprintf("\n %d | ", row)))
				middle.WriteString(h.ErrorCode(middleInner.String()))
				break
			}
			if b == '\n' {
				middle.WriteString(h.Bold(fmt.Sprintf("\n %d | ", row)))
				middle.WriteString(h.ErrorCode(middleInner.String()))
				middleInner.Reset()
				row++
			}
		}
		for i < len(input)-1 {
			endInner.WriteByte(b)
			i++
			b = input[i]
			if b == '\n' {
				end.WriteString(h.Code(endInner.String()))
				endInner.Reset()
				row++
			}
		}
		// Ensure that we don't get two
		// lines with the same row number:
		//
		// 1 | ...
		// 1 | ...
		//
		if row == startRow {
			row++
		}
		end.WriteString(h.Bold(fmt.Sprintf("\n %d | ", row)))
	}
	var errline strings.Builder
	errline.WriteString(start.String())
	errline.WriteString(middle.String())
	errline.WriteString(end.String())
	return fmt.Sprintf("%s\n\n%s", errline.String(), e.Msg)
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
