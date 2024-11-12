package lexer

import (
	"fmt"
	"strconv"
	"unicode"

	e "github.com/fholmqvist/remlisp/err"
	h "github.com/fholmqvist/remlisp/highlight"
	tk "github.com/fholmqvist/remlisp/token"
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
	for l.inRange() && isIdentBody(l.ch) {
		line = append(line, l.ch)
		l.step()
	}
	s := string(line)
	if s == "true" || s == "false" {
		return tk.Bool{
			V: s == "true",
			P: l.Pos(),
		}, nil
	}
	if s == "nil" {
		return tk.Nil{P: l.Pos()}, nil
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

func (l *Lexer) step() {
	l.i++
	if l.i >= len(l.input) {
		l.ch = 0
		return
	}
	l.ch = l.input[l.i]
}

func (l Lexer) inRange() bool {
	return l.i < len(l.input)
}

func (l Lexer) peek() byte {
	if l.i+1 >= len(l.input) {
		return 0
	}
	return l.input[l.i+1]
}

func (l Lexer) Pos() tk.Position {
	return tk.NewPos(l.oldi, l.i)
}

func isNumber(b, b2 byte) bool {
	if b == '-' && unicode.IsNumber(rune(b2)) {
		return true
	}
	return unicode.IsNumber(rune(b))
}

func isSpace(b byte) bool {
	return b == ' '
}

func isNewLine(b byte) bool {
	return b == '\n'
}

func isIdent(b byte) bool {
	return b == '_' || unicode.IsLetter(rune(b))
}

func isIdentBody(b byte) bool {
	return isIdent(b) && !isDelimiter(b) || isDot(b) ||
		isMinus(b) || isQuestionMark(b) || isExclamationMark(b)
}

func isQuestionMark(b byte) bool {
	return b == '?'
}

func isExclamationMark(b byte) bool {
	return b == '!'
}

func isString(b byte) bool {
	return b == '"'
}

func isComment(b byte) bool {
	return b == ';'
}

func isOperator(b byte) bool {
	switch b {
	case '+', '-', '*', '/', '%', '=', '<', '>', '!':
		return true
	default:
		return false
	}
}

func isComplexOperator(a, b byte) bool {
	switch {
	case a == '!' && b == '=':
		return true
	case a == '<' && b == '=':
		return true
	case a == '>' && b == '=':
		return true
	default:
		return false
	}
}

func isMinus(b byte) bool {
	return b == '-'
}

func isComma(b byte) bool {
	return b == ','
}

func isLeftParens(b byte) bool {
	return b == '('
}

func isRightParens(b byte) bool {
	return b == ')'
}

func isLeftBracket(b byte) bool {
	return b == '['
}

func isRightBracket(b byte) bool {
	return b == ']'
}

func isLeftBrace(b byte) bool {
	return b == '{'
}

func isRightBrace(b byte) bool {
	return b == '}'
}

func isDot(b byte) bool {
	return b == '.'
}

func isColon(b byte) bool {
	return b == ':'
}

func isAmpersand(b byte) bool {
	return b == '&'
}

func isQuote(b byte) bool {
	return b == '\''
}

func isQuasiquote(b byte) bool {
	return b == '`'
}

func isDelimiter(b byte) bool {
	switch b {
	case ' ', ',', ':', '\n', '\t', '[', ']', '(', ')', '{', '}':
		return true
	default:
		return false
	}
}
