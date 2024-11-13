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

func New(l *lexer.Lexer, p *parser.Parser, c *compiler.Compiler) *Expander {
	return &Expander{
		macros: []*ex.Macro{},
		lex:    l,
		prs:    p,
		com:    c,
	}
}

func (e *Expander) Expand(exprs []ex.Expr) ([]ex.Expr, *er.Error) {
	e.exprs = exprs
	e.forwardDeclareMacros()
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
		js, erre := e.com.Compile([]ex.Expr{expr.E})
		if erre != nil {
			return nil, erre
		}
		bb, err := exec.Command("deno", "eval", fmt.Sprintf("console.log(%s)", js)).Output()
		if err != nil {
			return nil, errFromStr("failed to execute unquote: %s", err.Error())
		}
		lisp, err := pp.FromJS(bb)
		if err != nil {
			return nil, errFromStr("failed to parse unquote: %s", err.Error())
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
			return nil, errFromStr("expected 1 expression, got %d", len(exprs))
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
			Msg:   fmt.Sprintf("expected %d arguments, got %d", len(m.Params.V), len(list.V)-1),
			Start: pos.Start,
			End:   pos.End,
		}
	}
	args := map[string]ex.Expr{}
	for i := range m.Params.V {
		args[m.Params.V[i].String()] = list.V[i+1]
	}
	switch body := m.Body.(type) {
	case *ex.List:
		return e.replaceArguments(body, args, false), nil
	case *ex.Quasiquote:
		nbody, ok := body.E.(*ex.List)
		if !ok {
			return body.E, nil
		}
		return e.replaceArguments(nbody, args, true), nil
	case *ex.Quote:
		return body, nil
	default:
		return body, nil
	}
}

func (e *Expander) replaceArguments(list *ex.List, args map[string]ex.Expr, quasi bool) *ex.List {
	// Remove reference semantics.
	nlist := ex.List{V: append([]ex.Expr{}, list.V...)}
	for i, expr := range nlist.V {
		switch expr := expr.(type) {
		case *ex.List:
			nlist.V[i] = e.replaceArguments(expr, args, quasi)
		default:
			if quasi {
				// Strip unquote.
				arg, ok := args[expr.String()[1:]]
				if !ok {
					continue
				}
				nlist.V[i] = arg
			} else {
				arg, ok := args[expr.String()]
				if !ok {
					continue
				}
				nlist.V[i] = arg
			}
		}
	}
	return &nlist
}

func (e *Expander) forwardDeclareMacros() {
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

func errFromStr(format string, args ...any) *er.Error {
	return &er.Error{Msg: fmt.Sprintf(format, args...)}
}
