package lexer

import (
	"fmt"
	"strconv"

	e "github.com/fholmqvist/remlisp/err"
	h "github.com/fholmqvist/remlisp/highlight"
	tk "github.com/fholmqvist/remlisp/token"
	"github.com/fholmqvist/remlisp/token/operator"
)

type Lexer struct {
	input string
	ch    byte

	i    int
	oldi int
}

func New() *Lexer {
	return &Lexer{}
}

func (l *Lexer) Lex(input []byte) ([]tk.Token, *e.Error) {
	l.input = string(input)
	l.ch = input[0]
	l.i = 0
	l.oldi = 0
	tokens := []tk.Token{}
	for l.inRange() {
		t, err := l.lex()
		if err != nil {
			return nil, err
		}
		if t != nil {
			tokens = append(tokens, t)
		}
	}
	return tokens, nil
}

func (l *Lexer) LexString(input string) ([]tk.Token, *e.Error) {
	return l.Lex([]byte(input))
}

func (l *Lexer) lex() (tk.Token, *e.Error) {
	if !l.inRange() {
		return nil, nil
	}
	l.oldi = l.i
	p := l.peek()
	switch {
	case isNumber(l.ch, p):
		return l.lexNumber()
	case isIdent(l.ch):
		return l.lexIdent()
	case isString(l.ch):
		return l.lexString()
	case isColon(l.ch):
		l.step()
		return l.lexAtom()
	case isComment(l.ch):
		for l.inRange() && l.ch != '\n' {
			l.step()
		}
		return l.lex()
	case isOperator(l.ch):
		return l.lexOperator()
	case isComma(l.ch):
		l.step()
		return tk.Comma{P: l.Pos()}, nil
	case isSpace(l.ch), isNewLine(l.ch):
		l.step()
		if !l.inRange() {
			return nil, nil
		}
		return l.lex()
	case isLeftParens(l.ch):
		l.step()
		return tk.LeftParen{P: l.Pos()}, nil
	case isRightParens(l.ch):
		l.step()
		return tk.RightParen{P: l.Pos()}, nil
	case isLeftBracket(l.ch):
		l.step()
		return tk.LeftBracket{P: l.Pos()}, nil
	case isRightBracket(l.ch):
		l.step()
		return tk.RightBracket{P: l.Pos()}, nil
	case isLeftBrace(l.ch):
		l.step()
		return tk.LeftBrace{P: l.Pos()}, nil
	case isRightBrace(l.ch):
		l.step()
		return tk.RightBrace{P: l.Pos()}, nil
	case isDot(l.ch):
		l.step()
		return tk.Dot{P: l.Pos()}, nil
	case isAmpersand(l.ch):
		l.step()
		return tk.Ampersand{P: l.Pos()}, nil
	case isQuote(l.ch):
		l.step()
		return tk.Quote{P: l.Pos()}, nil
	case isQuasiquote(l.ch):
		l.step()
		return tk.Quasiquote{P: l.Pos()}, nil
	case isAtSign(l.ch):
		l.step()
		return tk.AtSign{P: l.Pos()}, nil
	default:
		pos := l.Pos()
		return nil, e.FromPosition(pos, fmt.Sprintf("%s %s: %q",
			h.Bold(pos.String()), h.Red("unexpected character"), l.ch))
	}
}

func (l *Lexer) lexNumber() (tk.Token, *e.Error) {
	line := []byte{l.ch}
	l.step()
	float := false
	for l.inRange() && !isDelimiter(l.ch) {
		if l.ch == '.' {
			float = true
		}
		line = append(line, l.ch)
		l.step()
	}
	if float {
		f, err := strconv.ParseFloat(string(line), 64)
		if err != nil {
			return nil, &e.Error{
				Msg:   fmt.Sprintf("invalid number: %q", line),
				Start: l.oldi,
				End:   l.i,
			}
		}
		return tk.Float{
			V: f,
			P: l.Pos(),
		}, nil
	} else {
		i, err := strconv.Atoi(string(line))
		if err != nil {
			return nil, &e.Error{
				Msg:   fmt.Sprintf("invalid number: %q", line),
				Start: l.oldi,
				End:   l.i,
			}
		}
		return tk.Int{
			V: i,
			P: l.Pos(),
		}, nil
	}
}

func (l *Lexer) lexIdent() (tk.Token, *e.Error) {
	line := []byte{}
	for l.inRange() && !isDelimiter(l.ch) {
		line = append(line, l.ch)
		l.step()
	}
	s := string(line)
	if s == "true" || s == "false" {
		return tk.Bool{
			V: s == "true",
			P: l.Pos(),
		}, nil
	} else if s == "nil" {
		return tk.Nil{P: l.Pos()}, nil
	}
	if _, err := operator.From(s); err == nil {
		return tk.Operator{
			V: s,
			P: l.Pos(),
		}, nil
	}
	return tk.Identifier{
		V: s,
		P: l.Pos(),
	}, nil
}

func (l *Lexer) lexString() (tk.String, *e.Error) {
	line := []byte{}
	l.step()
	for l.inRange() && l.ch != '"' {
		line = append(line, l.ch)
		l.step()
	}
	l.step()
	return tk.String{
		V: string(line),
		P: l.Pos(),
	}, nil
}

func (l *Lexer) lexAtom() (tk.Token, *e.Error) {
	ident, err := l.lexIdent()
	if err != nil {
		return nil, err
	}
	return tk.Atom{
		V: ident.String(),
		P: tk.Position{
			Start: l.oldi,
			End:   l.i,
		},
	}, nil
}

func (l *Lexer) lexOperator() (tk.Token, *e.Error) {
	op := []byte{l.ch}
	l.step()
	if isComplexOperator(op[0], l.ch) {
		op = append(op, l.ch)
		l.step()
	}
	return tk.Operator{
		V: string(op),
		P: l.Pos(),
	}, nil
}
