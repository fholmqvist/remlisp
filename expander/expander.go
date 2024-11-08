package expander

import (
	"fmt"

	e "github.com/fholmqvist/remlisp/err"
	ex "github.com/fholmqvist/remlisp/expr"
	h "github.com/fholmqvist/remlisp/highlight"
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
}

func New(exprs []ex.Expr) *Expander {
	return &Expander{
		exprs:  exprs,
		macros: []*ex.Macro{},
	}
}

func (e *Expander) Expand() ([]ex.Expr, *e.Error) {
	e.predeclareMacros()
	for _, expr := range e.exprs {
		if err := e.expand(expr); err != nil {
			return nil, err
		}
	}
	return e.exprs, nil
}

func (e *Expander) expand(expr ex.Expr) *e.Error {
	// TODO: Find all recursive calls.
	switch expr := expr.(type) {
	case *ex.List:
		return e.expandCall(expr)
	}
	return nil
}

func (e *Expander) expandCall(list *ex.List) *e.Error {
	if len(list.V) == 0 {
		return nil
	}
	for _, expr := range list.V {
		call, ok := expr.(ex.Identifier)
		if !ok {
			if list2, ok := expr.(*ex.List); ok {
				if err := e.expandCall(list2); err != nil {
					return err
				}
			}
			continue
		}
		macro, ok := e.findMacro(call.V)
		if !ok {
			continue
		}
		if err := e.expandMacro(macro, list); err != nil {
			return err
		}
		e.logMacroExpansion(call.V)
	}
	return nil
}

func (e *Expander) findMacro(name string) (*ex.Macro, bool) {
	for _, m := range e.macros {
		if m.Name == name {
			return m, true
		}
	}
	return nil, false
}

func (e *Expander) expandMacro(_ *ex.Macro, _ *ex.List) *e.Error {
	return nil
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
