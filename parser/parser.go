package parser

import (
	"fmt"

	e "github.com/fholmqvist/remlisp/err"
	ex "github.com/fholmqvist/remlisp/expr"
	h "github.com/fholmqvist/remlisp/highlight"
	tk "github.com/fholmqvist/remlisp/token"
)

type Parser struct {
	tokens []tk.Token
	exprs  []ex.Expr

	i int
}

func New(tokens []tk.Token) (*Parser, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty tokens")
	}
	return &Parser{
		tokens: tokens,
		exprs:  []ex.Expr{},
		i:      0,
	}, nil
}

func (p *Parser) Parse() ([]ex.Expr, *e.Error) {
	for p.inRange() {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if expr == nil {
			continue
		}
		p.exprs = append(p.exprs, expr)
	}
	return p.exprs, nil
}

func (p *Parser) parseExpr() (ex.Expr, *e.Error) {
	next, err := p.next()
	if err != nil {
		return nil, err
	}
	switch t := next.(type) {
	case tk.Int:
		return p.parseInt(t)
	case tk.Float:
		return p.parseFloat(t)
	case tk.LeftParen:
		return p.parseList()
	default:
		return nil, p.errLastTokenType("unexpected token", next)
	}
}

func (p *Parser) parseInt(i tk.Int) (ex.Expr, *e.Error) {
	return nil, nil
}

func (p *Parser) parseFloat(f tk.Float) (ex.Expr, *e.Error) {
	return nil, nil
}

func (p *Parser) parseList() (ex.Expr, *e.Error) {
	return nil, nil
}

func (p *Parser) next() (tk.Token, *e.Error) {
	if !p.inRange() {
		return nil, p.errLastTokenType("unexpected end of input", nil)
	}
	t := p.tokens[p.i]
	p.i++
	return t, nil
}

func (p *Parser) eat(t tk.Token) *e.Error {
	if !p.inRange() {
		return e.FromToken(t, "unexpected end of input")
	}
	if fmt.Sprintf("%T", p.tokens[p.i]) != fmt.Sprintf("%T", t) {
		return e.FromToken(t, fmt.Sprintf("expected %q, got %q", t, p.tokens[p.i]))
	}
	p.i++
	return nil
}

func (p Parser) errLastTokenType(msg string, args any) *e.Error {
	return e.FromToken(p.tokens[p.i-1],
		fmt.Sprintf("%s: %s: %T",
			h.Bold(p.tokens[p.i-1].Pos().String()),
			h.Red(msg),
			args))
}

func (p Parser) inRange() bool {
	return p.i < len(p.tokens)
}
