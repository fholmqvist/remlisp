package compiler

import (
	"fmt"
	"os"

	e "github.com/fholmqvist/remlisp/err"
	"github.com/fholmqvist/remlisp/expander"
	h "github.com/fholmqvist/remlisp/highlight"
	"github.com/fholmqvist/remlisp/lexer"
	"github.com/fholmqvist/remlisp/parser"
	"github.com/fholmqvist/remlisp/print"
	"github.com/fholmqvist/remlisp/transpiler"
)

type Compiler struct {
	lex *lexer.Lexer
	prs *parser.Parser
	trn *transpiler.Transpiler

	print bool
}

func New(l *lexer.Lexer, p *parser.Parser, t *transpiler.Transpiler) *Compiler {
	return &Compiler{
		lex: l,
		prs: p,
		trn: t,
	}
}

func (c *Compiler) CompileFile(filename string, print bool, expander *expander.Expander) ([]byte, string, *e.Error) {
	c.print = print
	bb, err := os.ReadFile(filename)
	if err != nil {
		return bb, "", &e.Error{Msg: fmt.Sprintf("error reading file: %s", err)}
	}
	code, erre := c.Compile(bb, expander)
	if erre != nil {
		return bb, "", erre
	}
	return bb, code, nil
}

func (c *Compiler) Compile(bb []byte, expander *expander.Expander) (string, *e.Error) {
	tokens, err := c.lex.Lex(bb)
	if err != nil {
		return "", wrap("lexing", err)
	}
	if c.print {
		print.Tokens(tokens)
	}
	exprs, err := c.prs.Parse(tokens)
	if err != nil {
		return "", wrap("parse", err)
	}
	if c.print {
		print.Exprs(exprs)
	}
	if c.print {
		print.ExpanderHeader()
	}
	exprs, err = expander.Expand(exprs, c.print)
	if err != nil {
		return "", wrap("expansion", err)
	}
	if c.print {
		print.Line()
	}
	code, err := c.trn.Transpile(exprs)
	if err != nil {
		return "", wrap("compile", err)
	}
	if c.print {
		print.Code(code)
	}
	return code, nil
}

func wrap(msg string, err *e.Error) *e.Error {
	err.Msg = fmt.Sprintf("%s: %s", h.Bold(h.Red(msg+" error")), err.Msg)
	return err
}
