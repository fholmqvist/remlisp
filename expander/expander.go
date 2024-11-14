package expander

import (
	"fmt"
	"os/exec"
	"strings"

	er "github.com/fholmqvist/remlisp/err"
	ex "github.com/fholmqvist/remlisp/expr"
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
			return e.eval(exp)
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

func (e *Expander) eval(list *ex.List) (ex.Expr, *er.Error) {
	expanded, erre := e.expand(list)
	if erre != nil {
		return nil, erre
	}
	js, erre := e.trn.TranspileOne(expanded)
	if erre != nil {
		return nil, erre
	}
	bb, err := exec.Command("deno", "eval", fmt.Sprintf("console.log(%s)", js)).Output()
	if err != nil {
		return list, nil // errFromStr("failed to eval: %s", err.Error())
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
	if len(m.Params.V) != len(list.V)-1 && !m.Params.HasAmpersand() {
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
		nlist := e.replaceArguments(body, args)
		return e.eval(nlist)
	case *ex.Quasiquote:
		e.pushQuasi()
		defer e.popQuasi()
		switch nbody := body.E.(type) {
		case *ex.List:
			nlist := e.replaceArguments(nbody, args)
			return e.expandQuasiquoteInner(nlist)
		case *ex.Vec:
			nlist := e.replaceArguments(nbody.ToList(), args)
			expanded, err := e.expandQuasiquoteInner(nlist)
			if err != nil {
				return nil, err
			}
			return expanded.(*ex.List).ToVec(), nil
		default:
			nbody = e.replaceArgument(nbody, args)
			return e.expandQuasiquoteInner(nbody)
		}
	case *ex.Quote:
		n := e.replaceArgument(body.E, args)
		return e.expand(n)
	default:
		n := e.replaceArgument(body, args)
		return e.expand(n)
	}
}

func (e *Expander) replaceArguments(list *ex.List, args map[string]ex.Expr) *ex.List {
	// Remove reference semantics.
	nlist := ex.List{V: append([]ex.Expr{}, list.V...)}
	for i, expr := range nlist.V {
		nlist.V[i] = e.replaceArgument(expr, args)
	}
	return &nlist
}

func (e *Expander) replaceArgument(expr ex.Expr, args map[string]ex.Expr) ex.Expr {
	switch expr := expr.(type) {
	case *ex.List:
		return e.replaceArguments(expr, args)
	default:
		name := expr.String()
		if e.inQuasiquote() && strings.HasPrefix(name, ",") {
			// Strip unquote.
			name = name[1:]
		}
		arg, ok := args[name]
		if !ok {
			return expr
		}
		return arg
	}
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

func macroReplacementArgs(params *ex.Vec, args *ex.List) (map[string]ex.Expr, *er.Error) {
	nargs := map[string]ex.Expr{}
	for i := range params.V {
		arg := args.V[i+1]
		switch param := params.V[i].(type) {
		case *ex.Vec:
			switch arg := arg.(type) {
			case *ex.Vec:
				if len(arg.V) != len(param.V) {
					return nil, errFromStr("expected %d arguments, got %d",
						len(param.V), len(arg.V))
				}
				for j := range param.V {
					nargs[param.V[j].String()] = arg.V[j]
				}
			default:
				return nil, errFromStr("expected a nested vector of parameters, got %T", arg)
			}
		case *ex.VariableArg:
			// TODO:
			// nlist := &ex.List{V: make([]ex.Expr, len(args.V[i+1:]))}
			// for i, arg := range args.V[i+1:] {
			// 	nlist.V[i] = &ex.Quote{E: arg, P: arg.Pos()}
			// }
			nargs[param.V.V] = &ex.List{V: args.V[i+1:]}
			return nargs, nil
		default:
			nargs[param.String()] = arg
		}
	}
	return nargs, nil
}
