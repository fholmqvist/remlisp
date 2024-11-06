package expr

import (
	tk "github.com/fholmqvist/remlisp/token"
)

type Expr interface {
	Expr()
	String() string
	Pos() tk.Position
}
