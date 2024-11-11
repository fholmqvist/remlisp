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

func New() *Parser {
	return &Parser{
		exprs: []ex.Expr{},
		i:     0,
	}
}

func (p *Parser) Parse(tokens []tk.Token) ([]ex.Expr, *e.Error) {
	p.tokens = tokens
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
	case tk.Nil:
		return p.parseNil(t)
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
	case tk.LeftBracket:
		return p.parseVec()
	case tk.LeftBrace:
		return p.parseMap()
	case tk.Ampersand:
		return p.parseVariableArg(t.P)
	case tk.Dot:
		return p.parseDot(t)
	case tk.Quote:
		return p.parseQuote(t)
	case tk.Quasiquote:
		return p.parseQuasiquote(t)
	case tk.Comma:
		return p.parseUnquote(t)
	default:
		return nil, p.errLastTokenType("unexpected token", next)
	}
}

func (p *Parser) parseNil(n tk.Nil) (ex.Expr, *e.Error) {
	return ex.Nil{P: n.P}, nil
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
	hd := list.Head()
	if hd == nil {
		return list, nil
	}
	list.P = tk.Between(
		hd.Pos().BumpLeft(),
		list.Last().Pos().BumpRight(),
	)
	switch hd.String() {
	case "fn":
		return p.parseFn(list)
	case "if":
		return p.parseIf(list)
	case "while":
		return p.parseWhile(list)
	case "do":
		return p.parseDo(list)
	case "var":
		return p.parseVar(list)
	case "set":
		return p.parseSet(list)
	case "get":
		return p.parseGet(list)
	case ".":
		return p.parseDotList(list)
	case "macro":
		return p.parseMacro(list)
	default:
		return list, nil
	}
}

func (p *Parser) parseFn(list *ex.List) (ex.Expr, *e.Error) {
	var anonymous bool
	fn := list.Pop()
	name, actual, ok := list.PopIdentifier()
	if !ok {
		if _, ok := actual.(*ex.Vec); ok {
			anonymous = true
		} else {
			return nil, p.errLastTokenType("expected identifier", actual)
		}
	}
	params, actual, ok := list.PopVec()
	if !ok {
		return nil, p.errLastTokenType("expected parameters", actual)
	}
	body := list.Pop()
	if body == nil {
		return nil, p.errLastTokenType("expected body", body)
	}
	if anonymous {
		return &ex.AnonymousFn{
			Params: params,
			Body:   body,
			P:      tk.Between(fn.Pos().BumpLeft(), body.Pos().BumpRight()),
		}, nil
	} else {
		return &ex.Fn{
			Name:   name.V,
			Params: params,
			Body:   body,
			P:      tk.Between(fn.Pos().BumpLeft(), body.Pos().BumpRight()),
		}, nil
	}
}

func (p *Parser) parseIf(list *ex.List) (ex.Expr, *e.Error) {
	iff := list.Pop()
	cond := list.Pop()
	if cond == nil {
		return nil, p.errLastTokenType("expected condition", cond)
	}
	then := list.Pop()
	if then == nil {
		return nil, p.errLastTokenType("expected then", then)
	}
	els := list.Pop()
	if els == nil {
		return nil, p.errLastTokenType("expected else", els)
	}
	return &ex.If{
		Cond: cond,
		Then: then,
		Else: els,
		P:    tk.Between(iff.Pos().BumpLeft(), els.Pos().BumpRight()),
	}, nil
}

func (p *Parser) parseWhile(list *ex.List) (ex.Expr, *e.Error) {
	if len(list.V) != 3 {
		return nil, p.errGot(list, "while requires three expressions", list.String())
	}
	return list, nil
}

func (p *Parser) parseDo(list *ex.List) (ex.Expr, *e.Error) {
	if len(list.V) == 0 {
		return nil, p.errWas(list, "expected do", list)
	}
	if len(list.V) == 1 {
		return nil, p.errWas(list, "expected body for do", nil)
	}
	return list, nil
}

func (p *Parser) parseVar(list *ex.List) (ex.Expr, *e.Error) {
	_, actual, ok := list.PopIdentifier()
	if !ok {
		return nil, p.errLastTokenType("expected var", actual)
	}
	name, actual, ok := list.PopIdentifier()
	if !ok {
		return nil, p.errLastTokenType("expected identifier", actual)
	}
	value := list.Pop()
	if value == nil {
		return nil, p.errLastTokenType("expected value", value)
	}
	return &ex.Var{
		Name: name.V,
		V:    value,
		P:    list.P,
	}, nil
}

func (p *Parser) parseSet(list *ex.List) (ex.Expr, *e.Error) {
	_, actual, ok := list.PopIdentifier()
	if !ok {
		return nil, p.errLastTokenType("expected set", actual)
	}
	namee := list.Pop()
	if namee == nil {
		return nil, p.errLastTokenType("expected value", namee)
	}
	var name string
	switch n := namee.(type) {
	case *ex.Unquote:
		name = n.E.String()
	default:
		name = n.String()
	}
	expr := list.Pop()
	if expr == nil {
		return nil, p.errLastTokenType("expected value", expr)
	}
	return &ex.Set{
		Name: name,
		E:    expr,
		P:    list.P,
	}, nil
}

func (p *Parser) parseGet(list *ex.List) (ex.Expr, *e.Error) {
	get, actual, ok := list.PopIdentifier()
	if !ok || get.String() != "get" {
		return nil, p.errLastTokenType("expected get", actual)
	}
	e := list.Pop()
	if e == nil {
		return nil, p.errLastTokenType("expected value", e)
	}
	i := list.Pop()
	if i == nil {
		return nil, p.errLastTokenType("expected value", i)
	}
	return &ex.Get{
		E: e,
		I: i,
		P: list.P,
	}, nil
}

func (p *Parser) parseVec() (ex.Expr, *e.Error) {
	vec := &ex.Vec{}
	for p.inRange() && !p.is(tk.RightBracket{}) {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if expr == nil {
			continue
		}
		vec.Append(expr)
	}
	if err := p.eat(tk.RightBracket{}); err != nil {
		return nil, err
	}
	return vec, nil
}

func (p *Parser) parseDotList(list *ex.List) (ex.Expr, *e.Error) {
	if len(list.V) == 0 {
		return nil, p.errWas(list, "expected dot list", list)
	}
	if len(list.V) == 1 {
		return nil, p.errWas(list, "expected arguments for dot list", nil)
	}
	return list, nil
}

func (p *Parser) parseMap() (ex.Expr, *e.Error) {
	mp := &ex.Map{}
	for p.inRange() && !p.is(tk.RightBrace{}) {
		k, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		v, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		mp.AddKV(k, v)
	}
	if err := p.eat(tk.RightBrace{}); err != nil {
		return nil, err
	}
	return mp, nil
}

func (p *Parser) parseVariableArg(pos tk.Position) (ex.Expr, *e.Error) {
	arg, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	ident, ok := arg.(ex.Identifier)
	if !ok {
		return nil, p.errLastTokenType("expected identifier", arg)
	}
	return &ex.VariableArg{
		V: ident,
		P: tk.Between(pos, arg.Pos()),
	}, nil
}

func (p *Parser) parseDot(dot tk.Dot) (ex.Identifier, *e.Error) {
	return ex.Identifier{
		V: ".",
		P: dot.P,
	}, nil
}

func (p *Parser) parseQuote(q tk.Quote) (ex.Expr, *e.Error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &ex.Quote{
		E: expr,
		P: tk.Between(q.Pos(), expr.Pos()),
	}, nil
}

func (p *Parser) parseQuasiquote(q tk.Quasiquote) (ex.Expr, *e.Error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &ex.Quasiquote{
		E: expr,
		P: tk.Between(q.Pos(), expr.Pos()),
	}, nil
}

func (p *Parser) parseUnquote(c tk.Comma) (ex.Expr, *e.Error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &ex.Unquote{
		E: expr,
		P: tk.Between(c.Pos(), expr.Pos()),
	}, nil
}

func (p *Parser) parseMacro(list *ex.List) (ex.Expr, *e.Error) {
	m := list.Pop()
	if m == nil {
		return nil, p.errLastTokenType("expected macro", m)
	}
	name, actual, ok := list.PopIdentifier()
	if !ok {
		return nil, p.errLastTokenType("expected identifier", actual)
	}
	params, actual, ok := list.PopVec()
	if !ok {
		return nil, p.errLastTokenType("expected parameters", actual)
	}
	body := list.Pop()
	if body == nil {
		return nil, p.errLastTokenType("expected body", body)
	}
	return &ex.Macro{
		Name:   name.V,
		Params: params,
		Body:   body,
		P:      tk.Between(m.Pos().BumpLeft(), body.Pos().BumpRight()),
	}, nil
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
