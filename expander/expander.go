package expander

import (
	"fmt"
	"os/exec"

	"github.com/fholmqvist/remlisp/compiler"
	er "github.com/fholmqvist/remlisp/err"
	ex "github.com/fholmqvist/remlisp/expr"
	h "github.com/fholmqvist/remlisp/highlight"
	"github.com/fholmqvist/remlisp/lexer"
	"github.com/fholmqvist/remlisp/parser"
	"github.com/fholmqvist/remlisp/pp"
)

// ================
// IDEA
// ================
//
// Sneaky sneaky just pipe to Deno
// and replace call site with result.
//
// INPUT
//   (macro double-sum [x y]
//     `(+ (add ,x ,y) (add ,x ,y))`)
//
//   (double-sum 1 2)
//
// OUTPUT
//   // (macro double-sum [x y]
//   //   `(+ (add ,x ,y) (add ,x ,y))`)
//
//   add(1, 2) + add(1, 2)
//
// ================

type Expander struct {
	exprs  []ex.Expr
	macros []*ex.Macro

	printouts int

	lex *lexer.Lexer
	prs *parser.Parser
	com *compiler.Compiler
}

func New(exprs []ex.Expr) *Expander {
	return &Expander{
		exprs:  exprs,
		macros: []*ex.Macro{},
		lex:    lexer.New(),
		prs:    parser.New(),
		com:    compiler.New(),
	}
}

func (e *Expander) Expand() ([]ex.Expr, *er.Error) {
	e.predeclareMacros()
	for i, expr := range e.exprs {
		expanded, err := e.expand(expr)
		if err != nil {
			return nil, err
		}
		e.exprs[i] = expanded
	}
	return e.exprs, nil
}

func (e *Expander) expand(expr ex.Expr) (ex.Expr, *er.Error) {
	// TODO: Find all nested calls.
	switch expr := expr.(type) {
	case *ex.List:
		return e.expandCall(expr)
	case *ex.Quote:
		return expr.E, nil
	case *ex.Quasiquote:
		return e.expandQuasiquote(expr)
	}
	return expr, nil
}

func (e *Expander) expandCall(list *ex.List) (ex.Expr, *er.Error) {
	if len(list.V) == 0 {
		return list, nil
	}
	for i, expr := range list.V {
		switch expr := expr.(type) {
		case *ex.List:
			expanded, err := e.expandCall(expr)
			if err != nil {
				return nil, err
			}
			list.V[i] = expanded
		case ex.Identifier:
			macro, ok := e.findMacro(expr.V)
			if !ok {
				continue
			}
			expanded, err := e.expandMacro(macro, list)
			if err != nil {
				return nil, err
			}
			e.logMacroExpansion(expr.V)
			return expanded, nil
		}
	}
	return list, nil
}

func (e *Expander) expandQuasiquote(expr *ex.Quasiquote) (ex.Expr, *er.Error) {
	inner := expr.E
	expanded, err := e.expandQuasiquoteInner(inner)
	if err != nil {
		return nil, err
	}
	return expanded, nil
}

func (e *Expander) expandQuasiquoteInner(expr ex.Expr) (ex.Expr, *er.Error) {
	switch expr := expr.(type) {
	case *ex.Unquote:
		// TODO: This is very much a standin hack to
		//       demonstrate that this actually works.
		js, err := e.com.Compile([]ex.Expr{expr.E})
		if err != nil {
			return nil, &er.Error{
				Msg:   fmt.Sprintf("failed to compile unquote: %s", err.Error()),
				Start: expr.P.Start,
				End:   expr.P.End,
			}
		}
		bb, err := exec.Command("deno", "eval", fmt.Sprintf("console.log(%s)", js)).Output()
		if err != nil {
			return nil, &er.Error{
				Msg:   fmt.Sprintf("failed to execute unquote: %s", err.Error()),
				Start: expr.P.Start,
				End:   expr.P.End,
			}
		}
		lisp, err := pp.FromJS(bb)
		if err != nil {
			return nil, &er.Error{
				Msg:   fmt.Sprintf("failed to parse unquote: %s", err.Error()),
				Start: expr.P.Start,
				End:   expr.Pos().End,
			}
		}
		tokens, erre := e.lex.Lex([]byte(lisp))
		if erre != nil {
			return nil, erre
		}
		exprs, erre := e.prs.Parse(tokens)
		if erre != nil {
			return nil, erre
		}
		if len(exprs) != 1 {
			return nil, &er.Error{
				Msg:   fmt.Sprintf("expected 1 expression, got %d", len(exprs)),
				Start: expr.P.Start,
				End:   expr.Pos().End,
			}
		}
		return exprs[0], nil
	default:
		return e.expand(expr)
	}
}

func (e *Expander) findMacro(name string) (*ex.Macro, bool) {
	for _, m := range e.macros {
		if m.Name == name {
			return m, true
		}
	}
	return nil, false
}

func (e *Expander) expandMacro(m *ex.Macro, list *ex.List) (ex.Expr, *er.Error) {
	pos := list.P
	if len(m.Params.V) != len(list.V)-1 {
		return nil, &er.Error{
			Msg:   fmt.Sprintf("expected %d arguments, got %d", len(m.Params.V), len(list.V)),
			Start: pos.Start,
			End:   pos.End,
		}
	}
	args := map[string]ex.Expr{}
	for i := range m.Params.V {
		args[m.Params.V[i].String()] = list.V[i+1]
	}
	bls, ok := m.Body.(*ex.List)
	if !ok {
		return nil, &er.Error{
			Msg:   fmt.Sprintf("expected list, got %T", m.Body),
			Start: pos.Start,
			End:   pos.End,
		}
	}
	nbody := *bls
	for i, ex := range nbody.V {
		arg, ok := args[ex.String()]
		if !ok {
			continue
		}
		nbody.V[i] = arg
	}
	return &nbody, nil
}

func (e *Expander) predeclareMacros() {
	for _, expr := range e.exprs {
		if m, ok := expr.(*ex.Macro); ok {
			e.macros = append(e.macros, m)
			e.logMacro(m)
		}
	}
}

func (e *Expander) logMacro(m *ex.Macro) {
	num := fmt.Sprintf("%.4d", e.printouts)
	line := fmt.Sprintf("%s: %v", h.Bold("Read macro"), m.Name)
	fmt.Printf("%s | %s\n", h.Gray(num), line)
	e.printouts++
}

func (e *Expander) logMacroExpansion(name string) {
	num := fmt.Sprintf("%.4d", e.printouts)
	line := fmt.Sprintf("%s: %v", h.Bold("Expanded"), name)
	fmt.Printf("%s | %s\n", h.Gray(num), line)
	e.printouts++
}
