package compiler

import (
	"fmt"
	"strings"

	ex "github.com/fholmqvist/remlisp/expr"
)

type Compiler struct {
	exprs []ex.Expr
	i     int
}

func New(exprs []ex.Expr) (*Compiler, error) {
	if len(exprs) == 0 {
		return nil, fmt.Errorf("empty expressions")
	}
	return &Compiler{
		exprs: exprs,
		i:     0,
	}, nil
}

func (c *Compiler) Compile(exprs []ex.Expr) (string, error) {
	var s strings.Builder
	return s.String(), nil
}
