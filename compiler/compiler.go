package compiler

import (
	"fmt"
	"strings"

	"github.com/fholmqvist/remlisp/compiler/state"
	e "github.com/fholmqvist/remlisp/err"
	ex "github.com/fholmqvist/remlisp/expr"
	"github.com/fholmqvist/remlisp/token/operator"
)

type Compiler struct {
	exprs []ex.Expr
	i     int

	state    state.State
	oldstate []state.State
}

func New() *Compiler {
	return &Compiler{
		i:        0,
		state:    state.NORMAL,
		oldstate: []state.State{},
	}
}

func (c *Compiler) Compile(exprs []ex.Expr) (string, *e.Error) {
	c.exprs = exprs
	var s strings.Builder
	for _, e := range c.exprs {
		code, err := c.compile(e)
		if err != nil {
			return "", err
		}
		s.WriteString(code)
	}
	return s.String(), nil
}

func (c *Compiler) compile(expr ex.Expr) (string, *e.Error) {
	switch expr := expr.(type) {
	case ex.Nil:
		return "nil", nil
	case ex.Int:
		return fmt.Sprintf("%d", expr.V), nil
	case ex.Float:
		return fmt.Sprintf("%f", expr.V), nil
	case ex.Bool:
		return fmt.Sprintf("%t", expr.V), nil
	case ex.String:
		return fmt.Sprintf("%q", expr.V), nil
	case ex.Identifier:
		return fixName(expr.V), nil
	case ex.Atom:
		return fmt.Sprintf("%q", expr.String()), nil
	case *ex.List:
		return c.compileList(expr)
	case *ex.DotList:
		return c.compileDotList(expr)
	case *ex.Vec:
		return c.compileVec(expr)
	case *ex.Fn:
		return c.compileFn(expr)
	case *ex.AnonymousFn:
		return c.compileAnonymousFn(expr)
	case *ex.VariableArg:
		return c.compileVariableArg(expr)
	case *ex.If:
		return c.compileIf(expr)
	case *ex.While:
		return c.compileWhile(expr)
	case *ex.Do:
		return c.compileDo(expr)
	case *ex.Var:
		return c.compileVar(expr)
	case *ex.Set:
		return c.compileSet(expr)
	case *ex.Get:
		return c.compileGet(expr)
	case *ex.Map:
		return c.compileMap(expr)
	case *ex.Macro:
		return c.compileMacro(expr)
	case ex.Op:
		return "", e.FromPosition(expr.Pos(), fmt.Sprintf("misplaced operator: %q", expr))
	default:
		return "", e.FromPosition(expr.Pos(), fmt.Sprintf("unknown expression type: %T", expr))
	}
}

func (c *Compiler) compileList(e *ex.List) (string, *e.Error) {
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
		c.setState(state.NO_SEMICOLON)
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
		c.restoreState()
		if c.state == state.NO_SEMICOLON {
			s.WriteByte(')')
		} else {
			s.WriteString(");")
		}
		return s.String(), nil
	} else {
		s.WriteByte('[')
		for i, expr := range e.V {
			code, err := c.compile(expr)
			if err != nil {
				return "", err
			}
			s.WriteString(code)
			if i < len(e.V)-1 {
				s.WriteString(", ")
			}
		}
		s.WriteByte(']')
		return s.String(), nil
	}

}

func (c *Compiler) compileDotList(e *ex.DotList) (string, *e.Error) {
	var s strings.Builder
	c.setState(state.NO_SEMICOLON)
	defer c.restoreState()
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

func (c *Compiler) compileBinaryOperation(e *ex.List, op operator.Operator) (string, *e.Error) {
	c.setState(state.NO_SEMICOLON)
	defer c.restoreState()
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

func (c *Compiler) compileFn(fn *ex.Fn) (string, *e.Error) {
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

func (c *Compiler) compileAnonymousFn(fn *ex.AnonymousFn) (string, *e.Error) {
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

func (c *Compiler) compileIf(e *ex.If) (string, *e.Error) {
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

func (c *Compiler) compileWhile(e *ex.While) (string, *e.Error) {
	var s strings.Builder
	s.WriteString("(() => { ")
	cond, err := c.compile(e.Cond)
	if err != nil {
		return "", err
	}
	s.WriteString(fmt.Sprintf("while (%s) { ", cond))
	body, err := c.compile(e.Body)
	if err != nil {
		return "", err
	}
	s.WriteString(body)
	s.WriteString(" } })()")
	return s.String(), nil
}

func (c *Compiler) compileDo(e *ex.Do) (string, *e.Error) {
	c.setState(state.NO_SEMICOLON)
	defer c.restoreState()
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

func (c *Compiler) compileVar(e *ex.Var) (string, *e.Error) {
	name := fixName(e.Name)
	v, err := c.compile(e.V)
	if err != nil {
		return "", err
	}
	if c.state == state.NO_SEMICOLON {
		return fmt.Sprintf("let %s = %s", name, v), nil
	} else {
		return fmt.Sprintf("let %s = %s;", name, v), nil
	}
}

func (c *Compiler) compileSet(e *ex.Set) (string, *e.Error) {
	name := fixName(e.Name)
	code, err := c.compile(e.E)
	if err != nil {
		return "", err
	}
	if c.state == state.NO_SEMICOLON {
		return fmt.Sprintf("%s = %s", name, code), nil
	} else {
		return fmt.Sprintf("%s = %s;", name, code), nil
	}
}

func (c *Compiler) compileGet(e *ex.Get) (string, *e.Error) {
	ee, err := c.compile(e.E)
	if err != nil {
		return "", err
	}
	i, err := c.compile(e.I)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s[%s]", ee, i), nil
}

func (c *Compiler) compileMap(e *ex.Map) (string, *e.Error) {
	var s strings.Builder
	s.WriteString("({")
	for i, expr := range e.V {
		code, err := c.compile(expr)
		if err != nil {
			return "", err
		}
		s.WriteString(code)
		if i%2 == 0 {
			s.WriteByte(':')
		}
		if i < len(e.V)-1 {
			s.WriteByte(' ')
		}
	}
	s.WriteString("})")
	return s.String(), nil
}

func (c *Compiler) compileVec(e *ex.Vec) (string, *e.Error) {
	var s strings.Builder
	s.WriteByte('[')
	for i, expr := range e.V {
		code, err := c.compile(expr)
		if err != nil {
			return "", err
		}
		s.WriteString(code)
		if i < len(e.V)-1 {
			s.WriteString(", ")
		}
	}
	s.WriteByte(']')
	return s.String(), nil
}

func (c *Compiler) compileVariableArg(e *ex.VariableArg) (string, *e.Error) {
	arg, err := c.compile(e.V)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("...%s", arg), nil
}

func (c *Compiler) compileMacro(m *ex.Macro) (string, *e.Error) {
	lines := strings.Split(m.String(), "\n")
	return fmt.Sprintf("// %s\n\n", strings.Join(lines, "\n// ")), nil
}
