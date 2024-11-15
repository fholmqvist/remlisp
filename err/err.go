package err

import (
	"bytes"
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
		i      = lineAbove(input, e)
	)
	for ; i < len(input); i++ {
		b := input[i]
		var (
			startInner  strings.Builder
			middleInner strings.Builder
			endInner    strings.Builder
		)
		for i < e.Start {
			if b != '\n' {
				startInner.WriteByte(b)
			}
			i++
			if i >= len(input)-1 {
				break
			}
			b = input[i]
			if b == '\n' {
				start.WriteString(h.Bold(fmt.Sprintf("\n %.2d | ", row)))
				start.WriteString(h.Code(startInner.String()))
				startInner.Reset()
				row++
			}
		}
		if startInner.Len() > 0 {
			middle.WriteString(h.Bold(fmt.Sprintf("\n %.2d | ", row)))
			middle.WriteString(h.Code(startInner.String()))
		}
		for i < e.End {
			if b != '\n' {
				middleInner.WriteByte(b)
			}
			i++
			if i >= len(input)-1 {
				break
			}
			b = input[i]
			if i >= e.End {
				middle.WriteString(h.ErrorCode(middleInner.String()))
				break
			}
			if b == '\n' {
				middle.WriteString(h.Bold(fmt.Sprintf("\n %.2d | ", row)))
				middle.WriteString(h.ErrorCode(middleInner.String()))
				middleInner.Reset()
				row++
			}
		}
		for i < len(input)-1 {
			if b != '\n' {
				endInner.WriteByte(b)
			}
			i++
			if i >= len(input)-1 {
				break
			}
			b = input[i]
			if b == '\n' {
				end.WriteString(h.Code(endInner.String()))
				endInner.Reset()
				row++
				end.WriteString(h.Bold(fmt.Sprintf("\n %.2d | ", row)))
			}
		}
	}
	var errline strings.Builder
	errline.WriteString(start.String())
	errline.WriteString(middle.String())
	errline.WriteString(end.String())
	errstr := errline.String()
	if len(errstr) > 0 {
		return fmt.Sprintf("%s\n\n%s", errstr, e.Msg)
	} else {
		return fmt.Sprintf("\n%s", e.Msg)
	}
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

func lineAbove(input []byte, e Error) int {
	if len(input) < e.Start {
		return 0
	}
	var (
		last = 0
		curr = 0
	)
	for {
		best := bytes.Index(input[curr:e.Start], []byte("\n"))
		if best >= 0 {
			best += 1
			last = curr
			curr += best
		} else if best < 0 {
			break
		}
	}
	return last
}
