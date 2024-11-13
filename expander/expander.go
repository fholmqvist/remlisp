package expander

import (
	"fmt"
	"os/exec"

	er "github.com/fholmqvist/remlisp/err"
	ex "github.com/fholmqvist/remlisp/expr"
	h "github.com/fholmqvist/remlisp/highlight"
	"github.com/fholmqvist/remlisp/lexer"
	"github.com/fholmqvist/remlisp/parser"
	"github.com/fholmqvist/remlisp/pp"
	"github.com/fholmqvist/remlisp/transpiler"
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
	trn *transpiler.Transpiler

	quasi []struct{}

	print bool
}

func New(l *lexer.Lexer, p *parser.Parser, t *transpiler.Transpiler) *Expander {
	return &Expander{
		macros: []*ex.Macro{},
		lex:    l,
		prs:    p,
		trn:    t,
		quasi:  []struct{}{},
	}
}

func (e *Expander) Expand(exprs []ex.Expr, print bool) ([]ex.Expr, *er.Error) {
	e.print = print
	e.exprs = exprs
	e.printouts = 0
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
	case *ex.Vec:
		return e.expandVec(expr)
	case *ex.Quote:
		return expr.E, nil
	case *ex.Unquote:
		if !e.inQuasiquote() {
			return nil, &er.Error{
				Msg:   "unquote outside of quasiquote",
				Start: expr.P.Start,
				End:   expr.P.End,
			}
		} else {
			return expr.E, nil
		}
	case *ex.Quasiquote:
		return e.expandQuasiquote(expr)
	case *ex.Fn:
		paramse, err := e.expand(expr.Params)
		if err != nil {
			return nil, err
		}
		params, ok := paramse.(*ex.Vec)
		if !ok {
			return nil, &er.Error{
				Msg:   "expected a vector of parameters",
				Start: expr.Params.P.Start,
				End:   expr.Params.P.End,
			}
		}
		expr.Params = params
		body, err := e.expand(expr.Body)
		if err != nil {
			return nil, err
		}
		expr.Body = body
		return expr, nil
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
			if e.print {
				e.logMacroExpansion(macro.Name)
			}
			if i == 0 {
				return expanded, nil
			} else {
				list.V[i] = expanded
			}
		default:
			expanded, err := e.expand(expr)
			if err != nil {
				return nil, err
			}
			list.V[i] = expanded
		}
	}
	return list, nil
}

func (e *Expander) expandVec(vec *ex.Vec) (ex.Expr, *er.Error) {
	for i, expr := range vec.V {
		expanded, err := e.expand(expr)
		if err != nil {
			return nil, err
		}
		vec.V[i] = expanded
	}
	return vec, nil
}

func (e *Expander) expandQuasiquote(expr *ex.Quasiquote) (ex.Expr, *er.Error) {
	e.pushQuasi()
	defer e.popQuasi()
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
		switch exp := expr.E.(type) {
		case *ex.List:
			// TODO: This is very much a standin hack to
			//       demonstrate that this actually works.
			js, erre := e.trn.Transpile([]ex.Expr{expr.E})
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
			expanded, err := e.expand(exp)
			if err != nil {
				return nil, err
			}
			return expanded, nil
		}
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
	args, err := macroReplacementArgs(m.Params, list)
	if err != nil {
		return nil, err
	}
	switch body := m.Body.(type) {
	case *ex.List:
		return e.replaceArguments(body, args), nil
	case *ex.Quasiquote:
		nbody, ok := body.E.(*ex.List)
		if !ok {
			return body.E, nil
		}
		e.pushQuasi()
		defer e.popQuasi()
		nlist := e.replaceArguments(nbody, args)
		return e.expandQuasiquoteInner(nlist)
	case *ex.Quote:
		return body, nil
	default:
		return body, nil
	}
}

func (e *Expander) replaceArguments(list *ex.List, args map[string]ex.Expr) *ex.List {
	// Remove reference semantics.
	nlist := ex.List{V: append([]ex.Expr{}, list.V...)}
	for i, expr := range nlist.V {
		switch expr := expr.(type) {
		case *ex.List:
			nlist.V[i] = e.replaceArguments(expr, args)
		default:
			if e.inQuasiquote() {
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
			if e.print {
				e.logMacro(m)
			}
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
	if e.print {
		num := fmt.Sprintf("%.4d", e.printouts)
		line := fmt.Sprintf("%s: %v", h.Bold("Expanded"), name)
		fmt.Printf("%s | %s\n", h.Gray(num), line)
		e.printouts++
	}
}

func errFromStr(format string, args ...any) *er.Error {
	return &er.Error{Msg: fmt.Sprintf(format, args...)}
}

func macroReplacementArgs(params *ex.Vec, args *ex.List) (map[string]ex.Expr, *er.Error) {
	nargs := map[string]ex.Expr{}
	for i := range params.V {
		arg := args.V[i+1]
		switch param := params.V[i].(type) {
		case *ex.Vec:
			switch arg := arg.(type) {
			case *ex.Vec:
				if len(arg.V) != len(param.V) {
					return nil, errFromStr("expected %d arguments, got %d", len(param.V), len(arg.V))
				}
				for j := range param.V {
					nargs[param.V[j].String()] = arg.V[j]
				}
			default:
				return nil, errFromStr("expected a nested vector of parameters, got %T", arg)
			}
		default:
			nargs[params.V[i].String()] = arg
		}
	}
	return nargs, nil
}

func (e *Expander) inQuasiquote() bool {
	return len(e.quasi) > 0
}

func (e *Expander) pushQuasi() {
	e.quasi = append(e.quasi, struct{}{})
}

func (e *Expander) popQuasi() {
	e.quasi = e.quasi[:len(e.quasi)-1]
}
