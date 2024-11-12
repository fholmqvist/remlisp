package parser

import (
	"fmt"

	e "github.com/fholmqvist/remlisp/err"
	ex "github.com/fholmqvist/remlisp/expr"
	h "github.com/fholmqvist/remlisp/highlight"
	tk "github.com/fholmqvist/remlisp/token"
)

func (p *Parser) next() (tk.Token, *e.Error) {
	if !p.inRange() {
		return nil, p.errLastTokenType("unexpected end of input", nil)
	}
	t := p.tokens[p.i]
	p.i++
	return t, nil
}

func (p *Parser) is(t tk.Token) bool {
	return fmt.Sprintf("%T", p.tokens[p.i]) == fmt.Sprintf("%T", t)
}

func (p *Parser) eat(t tk.Token) *e.Error {
	if !p.inRange() {
		return e.FromToken(p.tokens[p.i-1], "unexpected end of input")
	}
	if fmt.Sprintf("%T", p.tokens[p.i]) != fmt.Sprintf("%T", t) {
		return e.FromToken(t, fmt.Sprintf("expected %q, got %q", t, p.tokens[p.i]))
	}
	p.i++
	return nil
}

func (p Parser) errLastTokenType(msg string, args any) *e.Error {
	t := p.tokens[p.i-1]
	pos := h.Bold(p.tokens[p.i-1].Pos().String())
	if args == nil {
		return e.FromToken(t, fmt.Sprintf("%s: %s: was %v",
			pos,
			h.Red(msg),
			args,
		))
	}
	if _, ok := args.(tk.Token); ok {
		return e.FromToken(t, fmt.Sprintf("%s: %s: %q",
			pos,
			h.Red(msg),
			args,
		))
	}
	return e.FromToken(t, fmt.Sprintf("%s: %s: was %T",
		pos,
		h.Red(msg),
		args,
	))
}

func (p Parser) errWas(expr ex.Expr, msg string, args any) *e.Error {
	pos := expr.Pos()
	postr := h.Bold(pos.String())
	if args == nil {
		return &e.Error{
			Msg: fmt.Sprintf("%s: %s: was %v",
				postr,
				h.Red(msg),
				args,
			),
			Start: pos.Start,
			End:   pos.End,
		}
	}
	if _, ok := args.(tk.Token); ok {
		return &e.Error{
			Msg: fmt.Sprintf("%s: %s: %q",
				postr,
				h.Red(msg),
				args,
			),
			Start: pos.Start,
			End:   pos.End,
		}
	}
	return &e.Error{
		Msg: fmt.Sprintf("%s: %s: was %T",
			postr,
			h.Red(msg),
			args,
		),
		Start: pos.Start,
		End:   pos.End,
	}
}

func (p Parser) errGot(expr ex.Expr, msg string, code string) *e.Error {
	pos := expr.Pos()
	postr := h.Bold(pos.String())
	return &e.Error{
		Msg: fmt.Sprintf("%s: %s: got: %v",
			postr,
			h.Red(msg),
			h.Code(code),
		),
		Start: pos.Start,
		End:   pos.End,
	}
}

func (p Parser) inRange() bool {
	return p.i < len(p.tokens)
}
