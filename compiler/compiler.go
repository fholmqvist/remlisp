package compiler

import (
	"fmt"
	"strings"

	ex "github.com/fholmqvist/remlisp/expr"
	"github.com/fholmqvist/remlisp/token/operator"
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
	for _, e := range exprs {
		code, err := c.compile(e)
		if err != nil {
			return "", err
		}
		s.WriteString(code)
	}
	return s.String(), nil
}

func (c *Compiler) compile(e ex.Expr) (string, error) {
	switch e := e.(type) {
	case ex.Int:
		return fmt.Sprintf("%d", e.V), nil
	case ex.Float:
		return fmt.Sprintf("%f", e.V), nil
	case ex.Bool:
		return fmt.Sprintf("%t", e.V), nil
	case ex.String:
		return fmt.Sprintf("%q", e.V), nil
	case ex.Identifier:
		return e.V, nil
	case ex.Op:
		return e.Op.String(), nil
	case *ex.List:
		return c.compileList(e)
	case *ex.Vec:
		return c.compileVec(e)
	default:
		return "", fmt.Errorf("unknown expression type: %T", e)
	}
}

func (c *Compiler) compileList(e *ex.List) (string, error) {
	var s strings.Builder
	if len(e.V) == 0 {
		return "()", nil
	}
	hd := e.V[0].String()
	switch hd {
	default:
		op, err := operator.From(hd)
		if err == nil {
			s.WriteByte('(')
			for i, expr := range e.V[1:] {
				code, err := c.compile(expr)
				if err != nil {
					return "", err
				}
				s.WriteString(code)
				if i < len(e.V)-2 {
					s.WriteString(fmt.Sprintf(" %s ", op.String()))
				}
			}
			s.WriteByte(')')
			return s.String(), nil
		}
		s.WriteByte('(')
		for i, expr := range e.V {
			code, err := c.compile(expr)
			if err != nil {
				return "", err
			}
			s.WriteString(code)
			if i < len(e.V)-1 {
				s.WriteByte(' ')
			}
		}
		s.WriteByte(')')
	}
	return s.String(), nil
}

func (c *Compiler) compileVec(e *ex.Vec) (string, error) {
	var s strings.Builder
	s.WriteByte('[')
	for i, expr := range e.V {
		code, err := c.compile(expr)
		if err != nil {
			return "", err
		}
		s.WriteString(code)
		if i < len(e.V)-1 {
			s.WriteByte(' ')
		}
	}
	s.WriteByte(']')
	return s.String(), nil
}
