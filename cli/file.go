package cli

import (
	"fmt"
	"os"

	"github.com/fholmqvist/remlisp/compiler"
	"github.com/fholmqvist/remlisp/expander"
	"github.com/fholmqvist/remlisp/lexer"
	"github.com/fholmqvist/remlisp/parser"
)

func compileFile(path string, print bool) string {
	bb, err := os.ReadFile(path)
	if err != nil {
		exit("reading input file", err)
	}
	lexer := lexer.New()
	tokens, erre := lexer.Lex(bb)
	if erre != nil {
		exite("lexing error", bb, erre)
	}
	if print {
		prettyPrintTokens(tokens)
	}
	parser := parser.New()
	exprs, erre := parser.Parse(tokens)
	if erre != nil {
		exite("parse error", bb, erre)
	}
	if print {
		prettyPrintExprs(exprs)
	}
	if print {
		printExpanderHeader()
	}
	exprs, erre = expander.New(exprs).Expand()
	if erre != nil {
		exite("expansion error", bb, erre)
	}
	fmt.Println()
	comp := compiler.New()
	code, erre := comp.Compile(exprs)
	if erre != nil {
		exite("compile error", bb, erre)
	}
	if print {
		prettyPrintCode(code)
	}
	return code
}

func createFile(filename, out string) error {
	return os.WriteFile(filename, []byte(out), os.ModePerm)
}
