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
	case ex.Nil:
		return "nil", nil
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
	case ex.Atom:
		return fmt.Sprintf("%q", e.String()), nil
	case *ex.List:
		return c.compileList(e)
	case *ex.DotList:
		return c.compileDotList(e)
	case *ex.Vec:
		return c.compileVec(e)
	case *ex.Fn:
		return c.compileFn(e)
	case *ex.AnonymousFn:
		return c.compileAnonymousFn(e)
	case *ex.VariableArg:
		return c.compileVariableArg(e)
	case *ex.If:
		return c.compileIf(e)
	case *ex.Do:
		return c.compileDo(e)
	case *ex.Var:
		return c.compileVar(e)
	case *ex.Set:
		return c.compileSet(e)
	case *ex.Get:
		return c.compileGet(e)
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
	op, err := operator.From(hd)
	if err == nil {
		return c.compileBinaryOperation(e, op)
	}
	if _, ok := e.V[0].(ex.Identifier); ok {
		s.WriteString(fixName(hd))
		s.WriteByte('(')
		rest := e.V[1:]
		for i, expr := range rest {
			code, err := c.compile(expr)
			if err != nil {
				return "", err
			}
			s.WriteString(code)
			if i < len(rest)-1 {
				s.WriteString(", ")
			}
		}
		s.WriteByte(')')
		return s.String(), nil
	} else {
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
		return s.String(), nil
	}

}

func (c *Compiler) compileDotList(e *ex.DotList) (string, error) {
	var s strings.Builder
	for i, expr := range e.V {
		code, err := c.compile(expr)
		if err != nil {
			return "", err
		}
		s.WriteString(code)
		if i < len(e.V)-1 {
			s.WriteByte('.')
		}
	}
	return s.String(), nil
}

func (c *Compiler) compileBinaryOperation(e *ex.List, op operator.Operator) (string, error) {
	opstr := op.String()
	if opstr == "=" {
		opstr = "=="
	}
	var s strings.Builder
	s.WriteByte('(')
	for i, expr := range e.V[1:] {
		code, err := c.compile(expr)
		if err != nil {
			return "", err
		}
		s.WriteString(code)
		if i < len(e.V)-2 {
			s.WriteString(fmt.Sprintf(" %s ", opstr))
		}
	}
	s.WriteByte(')')
	return s.String(), nil
}

func (c *Compiler) compileFn(fn *ex.Fn) (string, error) {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("const %s = (", fixName(fn.Name)))
	for i, p := range fn.Params.V {
		pstr, err := c.compile(p)
		if err != nil {
			return "", err
		}
		s.WriteString(pstr)
		if i < len(fn.Params.V)-1 {
			s.WriteString(", ")
		}
	}
	s.WriteString(") => ")
	body, err := c.compile(fn.Body)
	if err != nil {
		return "", err
	}
	s.WriteString(body)
	s.WriteString("\n\n")
	return s.String(), nil
}

func (c *Compiler) compileAnonymousFn(fn *ex.AnonymousFn) (string, error) {
	var s strings.Builder
	s.WriteByte('(')
	for i, p := range fn.Params.V {
		pstr, err := c.compile(p)
		if err != nil {
			return "", err
		}
		s.WriteString(pstr)
		if i < len(fn.Params.V)-1 {
			s.WriteString(", ")
		}
	}
	s.WriteString(") => ")
	body, err := c.compile(fn.Body)
	if err != nil {
		return "", err
	}
	s.WriteString(body)
	return s.String(), nil
}

func (c *Compiler) compileIf(e *ex.If) (string, error) {
	var s strings.Builder
	s.WriteString("(() => ")
	cond, err := c.compile(e.Cond)
	if err != nil {
		return "", err
	}
	s.WriteString(fmt.Sprintf("%s ? ", cond))
	then, err := c.compile(e.Then)
	if err != nil {
		return "", err
	}
	s.WriteString(then)
	s.WriteString(" : ")
	els, err := c.compile(e.Else)
	if err != nil {
		return "", err
	}
	s.WriteString(els)
	s.WriteString(")()")
	return s.String(), nil
}

func (c *Compiler) compileDo(e *ex.Do) (string, error) {
	var s strings.Builder
	s.WriteString("(() => { ")
	for i, expr := range e.V {
		code, err := c.compile(expr)
		if err != nil {
			return "", err
		}
		if i == len(e.V)-1 {
			s.WriteString("return ")
		}
		s.WriteString(code)
		s.WriteString("; ")
	}
	s.WriteString("})()")
	return s.String(), nil
}

func (c *Compiler) compileVar(e *ex.Var) (string, error) {
	return fmt.Sprintf("let %s = %s", fixName(e.Name), e.V), nil
}

func (c *Compiler) compileSet(e *ex.Set) (string, error) {
	code, err := c.compile(e.E)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s = %s", fixName(e.Name), code), nil
}

func (c *Compiler) compileGet(e *ex.Get) (string, error) {
	code, err := c.compile(e.E)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s[%s]", fixName(e.Name), code), nil
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

func (c *Compiler) compileVariableArg(e *ex.VariableArg) (string, error) {
	arg, err := c.compile(e.V)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("...%s", arg), nil
}

func fixName(s string) string {
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, "?", "P")
	s = strings.ReplaceAll(s, "!", "Ex")
	return s
}
