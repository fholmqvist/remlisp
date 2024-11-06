package parser

import (
	"fmt"

	e "github.com/fholmqvist/remlisp/err"
	ex "github.com/fholmqvist/remlisp/expr"
	h "github.com/fholmqvist/remlisp/highlight"
	tk "github.com/fholmqvist/remlisp/token"
	"github.com/fholmqvist/remlisp/token/operator"
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
	case tk.Bool:
		return p.parseBool(t)
	case tk.String:
		return p.parseString(t)
	case tk.Identifier:
		return p.parseIdentifier(t)
	case tk.Atom:
		return p.parseAtom(t)
	case tk.Operator:
		return p.parseOperator(t)
	case tk.LeftParen:
		return p.parseList()
	default:
		return nil, p.errLastTokenType("unexpected token", next)
	}
}

func (p *Parser) parseInt(i tk.Int) (ex.Expr, *e.Error) {
	return ex.Int{V: i.V, P: i.P}, nil
}

func (p *Parser) parseFloat(f tk.Float) (ex.Expr, *e.Error) {
	return ex.Float{V: f.V, P: f.P}, nil
}

func (p *Parser) parseOperator(o tk.Operator) (ex.Expr, *e.Error) {
	op, err := operator.From(o.V)
	if err != nil {
		return nil, e.FromToken(o, err.Error())
	}
	return ex.Op{
		Op: op,
		P:  o.P,
	}, nil
}

func (p *Parser) parseBool(b tk.Bool) (ex.Expr, *e.Error) {
	return ex.Bool{V: b.V, P: b.P}, nil
}

func (p *Parser) parseString(s tk.String) (ex.Expr, *e.Error) {
	return ex.String{V: s.V, P: s.P}, nil
}

func (p *Parser) parseIdentifier(i tk.Identifier) (ex.Expr, *e.Error) {
	return ex.Identifier{V: i.V, P: i.P}, nil
}

func (p *Parser) parseAtom(a tk.Atom) (ex.Expr, *e.Error) {
	return ex.Atom{V: a.V, P: a.P}, nil
}

func (p *Parser) parseList() (ex.Expr, *e.Error) {
	list := &ex.List{}
	for p.inRange() && !p.is(tk.RightParen{}) {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if expr == nil {
			continue
		}
		list.Append(expr)
	}
	if err := p.eat(tk.RightParen{}); err != nil {
		return nil, err
	}
	return list, nil
}

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
