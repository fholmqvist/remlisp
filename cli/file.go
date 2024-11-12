package cli

import (
	"fmt"
	"os"

	"github.com/fholmqvist/remlisp/compiler"
	"github.com/fholmqvist/remlisp/expander"
	"github.com/fholmqvist/remlisp/lexer"
	"github.com/fholmqvist/remlisp/parser"
)

func compileFile(path string, print bool, lexer *lexer.Lexer, parser *parser.Parser,
	expander *expander.Expander, compiler *compiler.Compiler,
) string {
	bb, err := os.ReadFile(path)
	if err != nil {
		exit("reading input file", err)
	}
	tokens, erre := lexer.Lex(bb)
	if erre != nil {
		exite("lexing error", bb, erre)
	}
	if print {
		prettyPrintTokens(tokens)
	}
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
	exprs, erre = expander.Expand(exprs)
	if erre != nil {
		exite("expansion error", bb, erre)
	}
	fmt.Println()
	code, erre := compiler.Compile(exprs)
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
