package expr

import (
	"fmt"
	"strings"

	tk "github.com/fholmqvist/remlisp/token"
	"github.com/fholmqvist/remlisp/token/operator"
)

type Expr interface {
	Expr()
	String() string
	Pos() tk.Position
}

type Nil struct {
	P tk.Position
}

func (Nil) Expr() {}

func (Nil) String() string {
	return "nil"
}

func (Nil) Pos() tk.Position {
	return tk.Position{}
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

type Bool struct {
	V bool
	P tk.Position
}

func (Bool) Expr() {}

func (b Bool) String() string {
	return fmt.Sprintf("%t", b.V)
}

func (b Bool) Pos() tk.Position {
	return b.P
}

type String struct {
	V string
	P tk.Position
}

func (String) Expr() {}

func (s String) String() string {
	return fmt.Sprintf("%q", s.V)
}

func (s String) Pos() tk.Position {
	return s.P
}

type Identifier struct {
	V string
	P tk.Position
}

func (Identifier) Expr() {}

func (i Identifier) String() string {
	return i.V
}

func (i Identifier) Pos() tk.Position {
	return i.P
}

type Atom struct {
	V string
	P tk.Position
}

func (Atom) Expr() {}

func (a Atom) String() string {
	return fmt.Sprintf(":%s", a.V)
}

func (a Atom) Pos() tk.Position {
	return a.P
}

type Op struct {
	Op operator.Operator
	P  tk.Position
}

func (Op) Expr() {}

func (o Op) String() string {
	return o.Op.String()
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

func (l *List) Head() Expr {
	if len(l.V) == 0 {
		return nil
	}
	return l.V[0]
}

func (l *List) Last() Expr {
	if len(l.V) == 0 {
		return nil
	}
	return l.V[len(l.V)-1]
}

func (l *List) Pop() Expr {
	if len(l.V) == 0 {
		return nil
	}
	hd := l.V[0]
	l.V = l.V[1:]
	return hd
}

func (l *List) PopIdentifier() (Identifier, Expr, bool) {
	if len(l.V) == 0 {
		return Identifier{}, nil, false
	}
	hd := l.V[0]
	id, ok := hd.(Identifier)
	if ok {
		l.V = l.V[1:]
	}
	return id, hd, ok
}

func (l *List) PopVec() (*Vec, Expr, bool) {
	if len(l.V) == 0 {
		return nil, nil, false
	}
	hd := l.V[0]
	v, ok := hd.(*Vec)
	if ok {
		l.V = l.V[1:]
	}
	return v, hd, ok
}

func (l *List) Append(e Expr) {
	l.V = append(l.V, e)
}

func (l List) ToVec() *Vec {
	v := &Vec{P: l.P}
	for _, e := range l.V {
		v.Append(e)
	}
	return v
}

type Vec struct {
	V []Expr
	P tk.Position
}

func (Vec) Expr() {}

func (v Vec) String() string {
	var s strings.Builder
	s.WriteByte('[')
	for i, e := range v.V {
		if i > 0 {
			s.WriteByte(' ')
		}
		s.WriteString(e.String())
	}
	s.WriteByte(']')
	return s.String()
}

func (v Vec) Pos() tk.Position {
	return v.P
}

func (v *Vec) Append(e Expr) {
	v.V = append(v.V, e)
}

func (v Vec) ToList() *List {
	l := &List{P: v.P}
	for _, e := range v.V {
		l.Append(e)
	}
	return l
}

type Map struct {
	V []Expr
	P tk.Position
}

func (Map) Expr() {}

func (m Map) String() string {
	var s strings.Builder
	s.WriteByte('{')
	for i, e := range m.V {
		if i > 0 {
			s.WriteByte(' ')
		}
		s.WriteString(e.String())
	}
	s.WriteByte('}')
	return s.String()
}

func (m Map) Pos() tk.Position {
	return m.P
}

func (m *Map) AddKV(k, v Expr) {
	m.V = append(m.V, k, v)
}

type Fn struct {
	Name      string
	Params    *Vec
	DocString string
	Body      Expr
	P         tk.Position
}

func (Fn) Expr() {}

func (f Fn) String() string {
	var s strings.Builder
	s.WriteString("(fn ")
	s.WriteString(f.Name)
	s.WriteString(" ")
	s.WriteString(f.Params.String())
	s.WriteString(" ")
	if f.DocString != "" {
		s.WriteString(fmt.Sprintf("%s ", f.DocString))
	}
	s.WriteString(f.Body.String())
	s.WriteByte(')')
	return s.String()
}

func (f Fn) Pos() tk.Position {
	return f.P
}

type AnonymousFn struct {
	Params *Vec
	Body   Expr
	P      tk.Position
}

func (AnonymousFn) Expr() {}

func (f AnonymousFn) String() string {
	var s strings.Builder
	s.WriteString("(fn ")
	s.WriteString(f.Params.String())
	s.WriteString(" ")
	s.WriteString(f.Body.String())
	s.WriteByte(')')
	return s.String()
}

func (f AnonymousFn) Pos() tk.Position {
	return f.P
}

type VariableArg struct {
	V Identifier
	P tk.Position
}

func (VariableArg) Expr() {}

func (v VariableArg) String() string {
	return fmt.Sprintf("& %s", v.V.String())
}

func (v VariableArg) Pos() tk.Position {
	return v.P
}

type Quote struct {
	E Expr
	P tk.Position
}

func (Quote) Expr() {}

func (q Quote) String() string {
	var st strings.Builder
	st.WriteByte('\'')
	st.WriteString(q.E.String())
	return st.String()
}

func (q Quote) Pos() tk.Position {
	return q.P
}

type Macro struct {
	Name   string
	Params *Vec
	Body   Expr
	P      tk.Position
}

func (Macro) Expr() {}

func (m Macro) String() string {
	var st strings.Builder
	st.WriteString("(macro ")
	st.WriteString(m.Name)
	st.WriteString(" ")
	st.WriteString(m.Params.String())
	st.WriteString(" ")
	st.WriteString(m.Body.String())
	st.WriteByte(')')
	return st.String()
}

func (m Macro) Pos() tk.Position {
	return m.P
}

type Quasiquote struct {
	E Expr
	P tk.Position
}

func (Quasiquote) Expr() {}

func (q Quasiquote) String() string {
	var st strings.Builder
	st.WriteString("`")
	st.WriteString(q.E.String())
	return st.String()
}

func (q Quasiquote) Pos() tk.Position {
	return q.P
}

type Unquote struct {
	E Expr
	P tk.Position
}

func (Unquote) Expr() {}

func (u Unquote) String() string {
	var st strings.Builder
	st.WriteString(",")
	st.WriteString(u.E.String())
	return st.String()
}

func (u Unquote) Pos() tk.Position {
	return u.P
}
