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
	c.i = 0
	c.state = state.NORMAL
	c.oldstate = []state.State{}
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
	case *ex.Vec:
		return c.compileVec(expr)
	case *ex.Fn:
		return c.compileFn(expr)
	case *ex.AnonymousFn:
		return c.compileAnonymousFn(expr)
	case *ex.VariableArg:
		return c.compileVariableArg(expr)
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

func (c *Compiler) compileList(list *ex.List) (string, *e.Error) {
	if len(list.V) == 0 {
		return "()", nil
	}
	head := list.V[0].String()
	op, err := operator.From(head)
	if err == nil {
		return c.compileBinaryOperation(list, op)
	}
	switch head {
	case "do":
		return c.compileDo(list)
	case "var":
		return c.compileVar(list)
	case "set":
		return c.compileSet(list)
	case "get":
		return c.compileGet(list)
	case "if":
		return c.compileIf(list)
	case "while":
		return c.compileWhile(list)
	case ".":
		return c.compileDotList(list)
	default:
		return c.compileListRaw(list, head)
	}
}

func (c *Compiler) compileListRaw(list *ex.List, head string) (string, *e.Error) {
	var s strings.Builder
	if _, ok := list.V[0].(ex.Identifier); ok {
		s.WriteString(fixName(head))
		s.WriteByte('(')
		c.setState(state.NO_SEMICOLON)
		rest := list.V[1:]
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
		for i, expr := range list.V {
			code, err := c.compile(expr)
			if err != nil {
				return "", err
			}
			s.WriteString(code)
			if i < len(list.V)-1 {
				s.WriteString(", ")
			}
		}
		s.WriteByte(']')
		return s.String(), nil
	}
}

func (c *Compiler) compileDotList(list *ex.List) (string, *e.Error) {
	var s strings.Builder
	c.setState(state.NO_SEMICOLON)
	defer c.restoreState()
	rest := list.V[1:]
	for i, expr := range rest {
		code, err := c.compile(expr)
		if err != nil {
			return "", err
		}
		s.WriteString(code)
		if i < len(rest)-1 {
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

func (c *Compiler) compileIf(list *ex.List) (string, *e.Error) {
	var s strings.Builder
	s.WriteString("(() => ")
	cond, err := c.compile(list.V[1])
	if err != nil {
		return "", err
	}
	s.WriteString(fmt.Sprintf("%s ? ", cond))
	then, err := c.compile(list.V[2])
	if err != nil {
		return "", err
	}
	s.WriteString(then)
	s.WriteString(" : ")
	els, err := c.compile(list.V[3])
	if err != nil {
		return "", err
	}
	s.WriteString(els)
	s.WriteString(")()")
	return s.String(), nil
}

func (c *Compiler) compileWhile(list *ex.List) (string, *e.Error) {
	var s strings.Builder
	s.WriteString("(() => { ")
	cond, err := c.compile(list.V[1])
	if err != nil {
		return "", err
	}
	s.WriteString(fmt.Sprintf("while (%s) { ", cond))
	body, err := c.compile(list.V[2])
	if err != nil {
		return "", err
	}
	s.WriteString(body)
	s.WriteString(" } })()")
	return s.String(), nil
}

func (c *Compiler) compileDo(list *ex.List) (string, *e.Error) {
	c.setState(state.NO_SEMICOLON)
	defer c.restoreState()
	var s strings.Builder
	s.WriteString("(() => { ")
	rest := list.V[1:]
	for i, expr := range rest {
		code, err := c.compile(expr)
		if err != nil {
			return "", err
		}
		if i == len(rest)-1 {
			s.WriteString("return ")
		}
		s.WriteString(code)
		s.WriteString("; ")
	}
	s.WriteString("})()")
	return s.String(), nil
}

func (c *Compiler) compileVar(list *ex.List) (string, *e.Error) {
	name := fixName(list.V[1].String())
	v, err := c.compile(list.V[2])
	if err != nil {
		return "", err
	}
	if c.state == state.NO_SEMICOLON {
		return fmt.Sprintf("let %s = %s", name, v), nil
	} else {
		return fmt.Sprintf("let %s = %s;", name, v), nil
	}
}

func (c *Compiler) compileSet(list *ex.List) (string, *e.Error) {
	name := fixName(list.V[1].String())
	code, err := c.compile(list.V[2])
	if err != nil {
		return "", err
	}
	if c.state == state.NO_SEMICOLON {
		return fmt.Sprintf("%s = %s", name, code), nil
	} else {
		return fmt.Sprintf("%s = %s;", name, code), nil
	}
}

func (c *Compiler) compileGet(list *ex.List) (string, *e.Error) {
	ee, err := c.compile(list.V[1])
	if err != nil {
		return "", err
	}
	i, err := c.compile(list.V[2])
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
