package transpiler

import (
	"fmt"
	"strings"

	e "github.com/fholmqvist/remlisp/err"
	ex "github.com/fholmqvist/remlisp/expr"
	"github.com/fholmqvist/remlisp/token/operator"
	"github.com/fholmqvist/remlisp/transpiler/state"
)

// TODO: Work on converting all statements to expressions.
//       Void becomes nil.

type Transpiler struct {
	exprs []ex.Expr
	i     int

	state    state.State
	oldstate []state.State
}

func New() *Transpiler {
	return &Transpiler{
		i:        0,
		state:    state.NORMAL,
		oldstate: []state.State{},
	}
}

func (t *Transpiler) Transpile(exprs []ex.Expr) (string, *e.Error) {
	t.exprs = exprs
	t.i = 0
	t.state = state.NORMAL
	t.oldstate = []state.State{}
	var s strings.Builder
	for _, e := range t.exprs {
		code, err := t.transpile(e)
		if err != nil {
			return "", err
		}
		s.WriteString(code)
	}
	return s.String(), nil
}

func (t *Transpiler) TranspileOne(expr ex.Expr) (string, *e.Error) {
	t.exprs = []ex.Expr{expr}
	t.i = 0
	t.state = state.NORMAL
	t.oldstate = []state.State{}
	var s strings.Builder
	code, err := t.transpile(expr)
	if err != nil {
		return "", err
	}
	s.WriteString(code)
	return s.String(), nil

}

func (t *Transpiler) transpile(expr ex.Expr) (string, *e.Error) {
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
		return fmt.Sprintf(`"%s"`, expr.V), nil
	case ex.Identifier:
		return fixName(expr.V), nil
	case ex.Atom:
		return fmt.Sprintf("%q", expr.String()), nil
	case *ex.List:
		return t.transpileList(expr)
	case *ex.Vec:
		return t.transpileVec(expr)
	case *ex.Fn:
		return t.transpileFn(expr)
	case *ex.AnonymousFn:
		return t.transpileAnonymousFn(expr)
	case *ex.VariableArg:
		return t.transpileVariableArg(expr)
	case *ex.Map:
		return t.transpileMap(expr)
	case *ex.Macro:
		return t.transpileMacro(expr)
	case *ex.Quote:
		return expr.E.String(), nil
	case ex.Op:
		return "", e.FromPosition(expr.Pos(), fmt.Sprintf("misplaced operator: %q", expr))
	default:
		return "", e.FromPosition(expr.Pos(), fmt.Sprintf("unknown expression type: %T", expr))
	}
}

func (t *Transpiler) transpileList(list *ex.List) (string, *e.Error) {
	if len(list.V) == 0 {
		return "()", nil
	}
	head := list.V[0].String()
	op, err := operator.From(head)
	if err == nil {
		return t.transpileBinaryOperation(list, op)
	}
	switch head {
	case "do":
		return t.transpileDo(list)
	case "var":
		return t.transpileVar(list)
	case "set":
		return t.transpileSet(list)
	case "get":
		return t.transpileGet(list)
	case "if":
		return t.transpileIf(list)
	case "while":
		return t.transpileWhile(list)
	case ".":
		return t.transpileDotList(list)
	default:
		return t.transpileListRaw(list, head)
	}
}

func (t *Transpiler) transpileListRaw(list *ex.List, head string) (string, *e.Error) {
	var s strings.Builder
	if _, ok := list.V[0].(ex.Identifier); ok {
		s.WriteString(fixName(head))
		s.WriteByte('(')
		t.setState(state.NO_SEMICOLON)
		rest := list.V[1:]
		for i, expr := range rest {
			code, err := t.transpile(expr)
			if err != nil {
				return "", err
			}
			s.WriteString(code)
			if i < len(rest)-1 {
				s.WriteString(", ")
			}
		}
		t.restoreState()
		if t.state == state.NO_SEMICOLON {
			s.WriteByte(')')
		} else {
			s.WriteString(");")
		}
		return s.String(), nil
	} else {
		s.WriteByte('[')
		for i, expr := range list.V {
			code, err := t.transpile(expr)
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

func (t *Transpiler) transpileDotList(list *ex.List) (string, *e.Error) {
	var s strings.Builder
	t.setState(state.NO_SEMICOLON)
	defer t.restoreState()
	rest := list.V[1:]
	for i, expr := range rest {
		code, err := t.transpile(expr)
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

func (t *Transpiler) transpileBinaryOperation(e *ex.List, op operator.Operator) (string, *e.Error) {
	t.setState(state.NO_SEMICOLON)
	defer t.restoreState()
	opstr := op.String()
	if opstr == "=" {
		opstr = "=="
	} else if opstr == "and" {
		opstr = "&&"
	} else if opstr == "or" {
		opstr = "||"
	}
	var s strings.Builder
	s.WriteByte('(')
	for i, expr := range e.V[1:] {
		code, err := t.transpile(expr)
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

func (t *Transpiler) transpileFn(fn *ex.Fn) (string, *e.Error) {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("function %s(", fixName(fn.Name)))
	for i, p := range fn.Params.V {
		pstr, err := t.transpile(p)
		if err != nil {
			return "", err
		}
		s.WriteString(pstr)
		if i < len(fn.Params.V)-1 {
			s.WriteString(", ")
		}
	}
	s.WriteString(") { return ")
	body, err := t.transpile(fn.Body)
	if err != nil {
		return "", err
	}
	s.WriteString(body)
	s.WriteString(" }\n\n")
	return s.String(), nil
}

func (t *Transpiler) transpileAnonymousFn(fn *ex.AnonymousFn) (string, *e.Error) {
	var s strings.Builder
	s.WriteByte('(')
	for i, p := range fn.Params.V {
		pstr, err := t.transpile(p)
		if err != nil {
			return "", err
		}
		s.WriteString(pstr)
		if i < len(fn.Params.V)-1 {
			s.WriteString(", ")
		}
	}
	s.WriteString(") => ")
	body, err := t.transpile(fn.Body)
	if err != nil {
		return "", err
	}
	s.WriteString(body)
	return s.String(), nil
}

func (t *Transpiler) transpileIf(list *ex.List) (string, *e.Error) {
	t.setState(state.NO_SEMICOLON)
	defer t.restoreState()
	var s strings.Builder
	s.WriteString("(() => ")
	cond, err := t.transpile(list.V[1])
	if err != nil {
		return "", err
	}
	s.WriteString(fmt.Sprintf("%s ? ", cond))
	then, err := t.transpile(list.V[2])
	if err != nil {
		return "", err
	}
	s.WriteString(then)
	s.WriteString(" : ")
	els, err := t.transpile(list.V[3])
	if err != nil {
		return "", err
	}
	s.WriteString(els)
	s.WriteString(")()")
	return s.String(), nil
}

func (t *Transpiler) transpileWhile(list *ex.List) (string, *e.Error) {
	var s strings.Builder
	s.WriteString("(() => { ")
	cond, err := t.transpile(list.V[1])
	if err != nil {
		return "", err
	}
	s.WriteString(fmt.Sprintf("while (%s) { ", cond))
	body, err := t.transpile(list.V[2])
	if err != nil {
		return "", err
	}
	s.WriteString(body)
	s.WriteString(" } })();")
	return s.String(), nil
}

func (t *Transpiler) transpileDo(list *ex.List) (string, *e.Error) {
	t.setState(state.NO_SEMICOLON)
	defer t.restoreState()
	var s strings.Builder
	s.WriteString("(() => { ")
	rest := list.V[1:]
	for i, expr := range rest {
		code, err := t.transpile(expr)
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
	if t.state != state.NO_SEMICOLON {
		s.WriteByte(';')
	}
	return s.String(), nil
}

func (t *Transpiler) transpileVar(list *ex.List) (string, *e.Error) {
	name := fixName(list.V[1].String())
	v, err := t.transpile(list.V[2])
	if err != nil {
		return "", err
	}
	if t.state == state.NO_SEMICOLON {
		return fmt.Sprintf("let %s = %s", name, v), nil
	} else {
		return fmt.Sprintf("let %s = %s;", name, v), nil
	}
}

func (t *Transpiler) transpileSet(list *ex.List) (string, *e.Error) {
	name, err := t.transpile(list.V[1])
	if err != nil {
		return "", err
	}
	name = fixName(name)
	code, err := t.transpile(list.V[2])
	if err != nil {
		return "", err
	}
	if t.state == state.NO_SEMICOLON {
		return fmt.Sprintf("%s = %s", name, code), nil
	} else {
		return fmt.Sprintf("%s = %s;", name, code), nil
	}
}

func (t *Transpiler) transpileGet(list *ex.List) (string, *e.Error) {
	ee, err := t.transpile(list.V[1])
	if err != nil {
		return "", err
	}
	i, err := t.transpile(list.V[2])
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s[%s]", ee, i), nil
}

func (t *Transpiler) transpileMap(e *ex.Map) (string, *e.Error) {
	var s strings.Builder
	s.WriteString("({")
	for i, expr := range e.V {
		code, err := t.transpile(expr)
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

func (t *Transpiler) transpileVec(e *ex.Vec) (string, *e.Error) {
	var s strings.Builder
	s.WriteByte('[')
	for i, expr := range e.V {
		code, err := t.transpile(expr)
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

func (t *Transpiler) transpileVariableArg(e *ex.VariableArg) (string, *e.Error) {
	arg, err := t.transpile(e.V)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("...%s", arg), nil
}

func (t *Transpiler) transpileMacro(m *ex.Macro) (string, *e.Error) {
	lines := strings.Split(m.String(), "\n")
	return fmt.Sprintf("// %s\n\n", strings.Join(lines, "\n// ")), nil
}
