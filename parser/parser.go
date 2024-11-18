package parser

import (
	"fmt"
	"slices"
	"strings"

	e "github.com/fholmqvist/remlisp/err"
	ex "github.com/fholmqvist/remlisp/expr"
	"github.com/fholmqvist/remlisp/lexer"
	"github.com/fholmqvist/remlisp/parser/state"
	tk "github.com/fholmqvist/remlisp/token"
	"github.com/fholmqvist/remlisp/token/operator"
)

type Parser struct {
	tokens []tk.Token
	exprs  []ex.Expr

	lex   *lexer.Lexer
	inner *Parser

	state    state.State
	oldstate []state.State

	i int
}

func New(lex *lexer.Lexer) *Parser {
	return nnew(nil, nnew(lex, nil))
}

// A parser with a nested parser.
//
//	parser := New(nil, New(lexer, nil))
//
// Parent parser doesn't need a lexer, nested does.
func nnew(lex *lexer.Lexer, inner *Parser) *Parser {
	return &Parser{
		exprs:    []ex.Expr{},
		lex:      lex,
		inner:    inner,
		state:    state.NORMAL,
		oldstate: []state.State{},
		i:        0,
	}
}

func (p *Parser) Parse(tokens []tk.Token) ([]ex.Expr, *e.Error) {
	p.exprs = []ex.Expr{}
	p.i = 0
	p.tokens = tokens
	for p.inRange() {
		expr, err := p.parse()
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

func (p *Parser) parse() (ex.Expr, *e.Error) {
	next, err := p.next()
	if err != nil {
		return nil, err
	}
	return p.parseExpr(next)
}

func (p *Parser) parseExpr(next tk.Token) (ex.Expr, *e.Error) {
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
	var statesSet int
	for p.inRange() && !p.is(tk.RightParen{}) {
		expr, err := p.parse()
		if err != nil {
			return nil, err
		}
		if expr == nil {
			continue
		}
		estr := expr.String()
		if estr == "->" || estr == "->>" {
			p.setState(state.THREADING)
			statesSet++
		}
		list.Append(expr)
	}
	for i := 0; i < statesSet; i++ {
		p.restoreState()
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
	case "match":
		return p.parseMatch(list)
	case "->":
		return p.parseThreadFirst(list)
	case "->>":
		return p.parseThreadLast(list)
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
	var docstring string
	if len(list.V) == 1 {
		docstring = body.String()
		body = list.Pop()
	}
	if anonymous {
		return &ex.AnonymousFn{
			Params: params,
			Body:   body,
			P:      tk.Between(fn.Pos().BumpLeft(), body.Pos().BumpRight()),
		}, nil
	} else {
		return &ex.Fn{
			Name:      name.V,
			Params:    params,
			DocString: docstring,
			Body:      body,
			P:         tk.Between(fn.Pos().BumpLeft(), body.Pos().BumpRight()),
		}, nil
	}
}

func (p *Parser) parseIf(list *ex.List) (ex.Expr, *e.Error) {
	if len(list.V) != 4 {
		return nil, p.errGot(list, "if requires three expressions", list.String())
	}
	return list, nil
}

func (p *Parser) parseWhile(list *ex.List) (ex.Expr, *e.Error) {
	if len(list.V) != 3 && p.state != state.THREADING {
		return nil, p.errGot(list, "while requires two expressions", list.String())
	}
	return list, nil
}

func (p *Parser) parseDo(list *ex.List) (ex.Expr, *e.Error) {
	if len(list.V) == 0 && p.state != state.THREADING {
		return nil, p.errWas(list, "expected do", list)
	}
	if len(list.V) == 1 && p.state != state.THREADING {
		return nil, p.errWas(list, "expected body for do", nil)
	}
	return list, nil
}

func (p *Parser) parseVar(list *ex.List) (ex.Expr, *e.Error) {
	if len(list.V) != 3 && p.state != state.THREADING {
		return nil, p.errGot(list, "var requires two expressions", list.String())
	}
	return list, nil
}

func (p *Parser) parseSet(list *ex.List) (ex.Expr, *e.Error) {
	if len(list.V) != 3 && p.state != state.THREADING {
		return nil, p.errGot(list, "set requires two expressions", list.String())
	}
	return list, nil
}

func (p *Parser) parseGet(list *ex.List) (ex.Expr, *e.Error) {
	if len(list.V) != 3 && p.state != state.THREADING {
		return nil, p.errGot(list, "get requires two expressions", list.String())
	}
	return list, nil
}

func (p *Parser) parseVec() (ex.Expr, *e.Error) {
	vec := &ex.Vec{}
	for p.inRange() && !p.is(tk.RightBracket{}) {
		expr, err := p.parse()
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
		k, err := p.parse()
		if err != nil {
			return nil, err
		}
		v, err := p.parse()
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
	arg, err := p.parse()
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
	expr, err := p.parse()
	if err != nil {
		return nil, err
	}
	return &ex.Quote{
		E: expr,
		P: tk.Between(q.Pos(), expr.Pos()),
	}, nil
}

func (p *Parser) parseQuasiquote(q tk.Quasiquote) (ex.Expr, *e.Error) {
	expr, err := p.parse()
	if err != nil {
		return nil, err
	}
	return &ex.Quasiquote{
		E: expr,
		P: tk.Between(q.Pos(), expr.Pos()),
	}, nil
}

func (p *Parser) parseUnquote(c tk.Comma) (ex.Expr, *e.Error) {
	t, err := p.next()
	if err != nil {
		return nil, err
	}
	switch next := t.(type) {
	case tk.AtSign:
		return p.parseUnquoteSplicing(next)
	default:
		expr, err := p.parseExpr(next)
		if err != nil {
			return nil, err
		}
		return &ex.Unquote{
			E: expr,
			P: tk.Between(c.Pos(), expr.Pos()),
		}, nil
	}
}

func (p *Parser) parseUnquoteSplicing(c tk.AtSign) (ex.Expr, *e.Error) {
	expr, err := p.parse()
	if err != nil {
		return nil, err
	}
	return &ex.UnquoteSplicing{
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

func (p *Parser) parseMatch(list *ex.List) (ex.Expr, *e.Error) {
	if len(list.V) == 0 {
		return nil, p.errWas(list, "expected match", list)
	}
	_ = list.Pop()
	cond := list.Pop()
	added := 0
	var s strings.Builder
	for len(list.V) > 0 {
		cmpe := list.Pop()
		body := list.Pop()
		switch cmp := cmpe.(type) {
		case *ex.List:
			cleancmp := cmp.String()
			cleancmp = strings.ReplaceAll(cleancmp, " _", " 0")
			cleancmp = strings.ReplaceAll(cleancmp, "_ ", "0 ")
			s.WriteString(fmt.Sprintf("(if (and (= (length %s) (length %s))", cond, cleancmp))
			if len(cmp.V) > 0 {
				for i, expr := range cmp.V {
					if expr.String() == "_" {
					} else {
						s.WriteString(fmt.Sprintf(" (= %s (get %s %d))", expr, cond, i))
					}
				}
			}
			s.WriteString(fmt.Sprintf(") %s ", body))
		case *ex.Vec:
			cleancmp := cmp.String()
			cleancmp = strings.ReplaceAll(cleancmp, " _", " 0")
			cleancmp = strings.ReplaceAll(cleancmp, "_ ", "0 ")
			s.WriteString(fmt.Sprintf("(if (and (= (length %s) (length %s))", cond, cleancmp))
			if len(cmp.V) > 0 {
				for i, expr := range cmp.V {
					if expr.String() == "_" {
					} else {
						s.WriteString(fmt.Sprintf(" (= %s (get %s %d))", expr, cond, i))
					}
				}
			}
			s.WriteString(fmt.Sprintf(") %s ", body))
		case ex.Atom:
			if cmp.V != "else" {
				return nil, p.errWas(cmp, "only :else atoms are supported as of now", cmp)
			}
			s.WriteString(body.String())
		default:
			return nil, p.errWas(cmp, "expected comparison", cmp)
		}
		added++
	}
	for i := 1; i < added; i++ {
		s.WriteByte(')')
	}
	tokens, err := p.inner.lex.LexString(s.String())
	if err != nil {
		return nil, err
	}
	nliste, err := p.inner.Parse(tokens)
	if err != nil {
		return nil, err
	}
	return nliste[0], nil
}

func (p *Parser) parseThreadFirst(list *ex.List) (ex.Expr, *e.Error) {
	if len(list.V) == 0 {
		return nil, p.errWas(list, "expected thread first", list)
	}
	_ = list.Pop()
	fst := list.Pop()
	snde := list.Pop()
	snd, ok := snde.(*ex.List)
	if !ok {
		return nil, p.errWas(snde, "expected list", snde)
	}
	if len(snd.V) > 1 {
		snd.V = slices.Insert(snd.V, 1, fst)
	} else {
		snd.Append(fst)
	}
	last := snd
	for len(list.V) > 0 {
		nexte := list.Pop()
		next, ok := nexte.(*ex.List)
		if !ok {
			return nil, p.errWas(nexte, "expected list", nexte)
		}
		if len(next.V) > 1 {
			next.V = slices.Insert(next.V, 1, ex.Expr(last))
		} else {
			next.Append(last)
		}
		last = next
	}
	return last, nil
}

func (p *Parser) parseThreadLast(list *ex.List) (ex.Expr, *e.Error) {
	if len(list.V) == 0 {
		return nil, p.errWas(list, "expected thread last", list)
	}
	_ = list.Pop()
	fst := list.Pop()
	snde := list.Pop()
	snd, ok := snde.(*ex.List)
	if !ok {
		return nil, p.errWas(snde, "expected list", snde)
	}
	snd.Append(fst)
	last := snd
	for len(list.V) > 0 {
		nexte := list.Pop()
		next, ok := nexte.(*ex.List)
		if !ok {
			return nil, p.errWas(nexte, "expected list", nexte)
		}
		next.Append(last)
		last = next
	}
	return last, nil
}
