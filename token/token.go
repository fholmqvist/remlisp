package token

import "fmt"

type Token interface {
	Token()
	String() string
	Pos() Position
}

type EOF struct{}

func (EOF) Token() {}

func (e EOF) String() string {
	return "EOF"
}

func (e EOF) Pos() Position {
	return Position{}
}

type Nil struct {
	P Position
}

func (Nil) Token() {}

func (n Nil) String() string {
	return "nil"
}

func (n Nil) Pos() Position {
	return n.P
}

type Int struct {
	V int
	P Position
}

func (Int) Token() {}

func (i Int) String() string {
	return fmt.Sprintf("%d", i.V)
}

func (i Int) Pos() Position {
	return i.P
}

type Float struct {
	V float64
	P Position
}

func (Float) Token() {}

func (f Float) String() string {
	return fmt.Sprintf("%.2f", f.V)
}

func (f Float) Pos() Position {
	return f.P
}

type Bool struct {
	V bool
	P Position
}

func (Bool) Token() {}

func (b Bool) String() string {
	return fmt.Sprintf("%t", b.V)
}

func (b Bool) Pos() Position {
	return b.P
}

type Identifier struct {
	V string
	P Position
}

func (Identifier) Token() {}

func (i Identifier) String() string {
	return i.V
}

func (i Identifier) Pos() Position {
	return i.P
}

type String struct {
	V string
	P Position
}

func (String) Token() {}

func (s String) String() string {
	return fmt.Sprintf("%q", s.V)
}

func (s String) Pos() Position {
	return s.P
}

type Atom struct {
	V string
	P Position
}

func (Atom) Token() {}

func (a Atom) String() string {
	return fmt.Sprintf(":%s", a.V)
}

func (a Atom) Pos() Position {
	return a.P
}

type Operator struct {
	V string
	P Position
}

func (Operator) Token() {}

func (o Operator) String() string {
	return o.V
}

func (o Operator) Pos() Position {
	return o.P
}

type Comma struct {
	P Position
}

func (Comma) Token() {}

func (c Comma) String() string {
	return ","
}

func (c Comma) Pos() Position {
	return c.P
}

type LeftParen struct {
	P Position
}

func (LeftParen) Token() {}

func (o LeftParen) String() string {
	return "("
}

func (o LeftParen) Pos() Position {
	return o.P
}

type RightParen struct {
	P Position
}

func (RightParen) Token() {}

func (c RightParen) String() string {
	return ")"
}

func (c RightParen) Pos() Position {
	return c.P
}

type LeftBracket struct {
	P Position
}

func (LeftBracket) Token() {}

func (o LeftBracket) String() string {
	return "["
}

func (o LeftBracket) Pos() Position {
	return o.P
}

type RightBracket struct {
	P Position
}

func (RightBracket) Token() {}

func (c RightBracket) String() string {
	return "]"
}

func (c RightBracket) Pos() Position {
	return c.P
}

type LeftBrace struct {
	P Position
}

func (LeftBrace) Token() {}

func (o LeftBrace) String() string {
	return "{"
}

func (o LeftBrace) Pos() Position {
	return o.P
}

type RightBrace struct {
	P Position
}

func (RightBrace) Token() {}

func (c RightBrace) String() string {
	return "}"
}

func (c RightBrace) Pos() Position {
	return c.P
}

type Dot struct {
	P Position
}

func (Dot) Token() {}

func (c Dot) String() string {
	return "."
}

func (c Dot) Pos() Position {
	return c.P
}

type Ampersand struct {
	P Position
}

func (Ampersand) Token() {}

func (c Ampersand) String() string {
	return "&"
}

func (c Ampersand) Pos() Position {
	return c.P
}

type Quote struct {
	P Position
}

func (Quote) Token() {}

func (c Quote) String() string {
	return "'"
}

func (c Quote) Pos() Position {
	return c.P
}
