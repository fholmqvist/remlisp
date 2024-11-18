package lexer

import (
	"unicode"

	tk "github.com/fholmqvist/remlisp/token"
)

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
	return !isDelimiter(b) && (unicode.IsLetter(rune(b)) || isUnderscore(b) || isDot(b) ||
		isMinus(b) || isRightArrow(b) || isQuestionMark(b) || isExclamationMark(b))
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

func isUnderscore(b byte) bool {
	return b == '_'
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

func isRightArrow(b byte) bool {
	return b == '>'
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

func isAtSign(b byte) bool {
	return b == '@'
}

func isDelimiter(b byte) bool {
	switch b {
	case ' ', ',', ':', '\n', '\t', '[', ']', '(', ')', '{', '}':
		return true
	default:
		return false
	}
}
