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
	lexer, err := lexer.New(bb)
	if err != nil {
		exit("instantiating lexer", err)
	}
	tokens, erre := lexer.Lex()
	if erre != nil {
		exite("lexing error", bb, erre)
	}
	if print {
		prettyPrintTokens(tokens)
	}
	parser, err := parser.New(tokens)
	if err != nil {
		exit("instantiating parser", err)
	}
	exprs, erre := parser.Parse()
	if erre != nil {
		exite("parse error", bb, erre)
	}
	if print {
		prettyPrintExprs(exprs)
	}
	printExpanderHeader()
	exprs, erre = expander.New(exprs).Expand()
	if erre != nil {
		exite("expansion error", bb, erre)
	}
	fmt.Println()
	comp, err := compiler.New(exprs)
	if err != nil {
		exit("error instantiating compiler", err)
	}
	code, err := comp.Compile()
	if err != nil {
		exit("compile error", err)
	}
	if print {
		prettyPrintCode(code)
	}
	return code
}

func createFile(filename, out string) error {
	return os.WriteFile(filename, []byte(out), os.ModePerm)
}
