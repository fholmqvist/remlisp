package expr

import (
	"fmt"
	"strings"

	tk "github.com/fholmqvist/remlisp/token"
)

type Expr interface {
	Expr()
	String() string
	Pos() tk.Position
}

type Int struct {
	V int
	P tk.Position
}

func (Int) Expr() {}

func (i Int) String() string {
	return fmt.Sprintf("%d", i.V)
}

func (i Int) Pos() tk.Position {
	return i.P
}

type Float struct {
	V float64
	P tk.Position
}

func (Float) Expr() {}

func (f Float) String() string {
	return fmt.Sprintf("%.2f", f.V)
}

func (f Float) Pos() tk.Position {
	return f.P
}

type Op struct {
	Op string
	P  tk.Position
}

func (Op) Expr() {}

func (o Op) String() string {
	return o.Op
}

func (o Op) Pos() tk.Position {
	return o.P
}

type List struct {
	V []Expr
	P tk.Position
}

func (List) Expr() {}

func (l List) String() string {
	var s strings.Builder
	s.WriteByte('(')
	for i, e := range l.V {
		if i > 0 {
			s.WriteByte(' ')
		}
		s.WriteString(e.String())
	}
	s.WriteByte(')')
	return s.String()
}

func (l List) Pos() tk.Position {
	return l.P
}

/*
	bool
	ident
	string
	atom
	nil
	quote
	args
	params
	fn
	list
	dotlist
	vector
	keyval
	map
	do
	if
	while
	var
	set
	get
	append
	macro
*/
