package main

import (
	"fmt"
	"os"

	e "github.com/fholmqvist/remlisp/err"
	ex "github.com/fholmqvist/remlisp/expr"
	h "github.com/fholmqvist/remlisp/highlight"
	"github.com/fholmqvist/remlisp/lexer"
	"github.com/fholmqvist/remlisp/parser"
	tk "github.com/fholmqvist/remlisp/token"
)

func main() {
	bb, err := os.ReadFile("input.rem")
	if err != nil {
		panic(err)
	}
	lexer, err := lexer.New(bb)
	if err != nil {
		exit("error instantiating lexer", err)
	}
	tokens, erre := lexer.Lex()
	if erre != nil {
		exite("lexing error", bb, erre)
	}
	prettyPrintTokens(tokens)
	parser, err := parser.New(tokens)
	if err != nil {
		exit("error instantiating parser", err)
	}
	exprs, erre := parser.Parse()
	if erre != nil {
		exite("parse error", bb, erre)
	}
	prettyPrintExprs(exprs)
}

func prettyPrintTokens(tokens []tk.Token) {
	fmt.Printf("\n%s\n", h.Bold("TOKENS ============="))
	if len(tokens) > 0 {
		for i, t := range tokens {
			num := fmt.Sprintf("%.4d", i)
			fmt.Printf("%s | %s (%T)\n",
				h.Gray(num), h.Code(t.String()), t)
		}
	} else {
		fmt.Println("<no tokens>")
	}
	printLine()
}

func prettyPrintExprs(exprs []ex.Expr) {
	fmt.Printf("%s\n", h.Bold("EXPRESSIONS ========"))
	if len(exprs) > 0 {
		for i, e := range exprs {
			num := fmt.Sprintf("%.4d", i)
			fmt.Printf("%s | %s (%T)\n",
				h.Gray(num), h.Code(e.String()), e)
		}
	} else {
		fmt.Println("<no expressions>")
	}
	printLine()
}

func printLine() {
	fmt.Printf("%s\n\n",
		"====================")
}

func exit(context string, err error) {
	fmt.Printf("%s: %s\n\n", h.Bold(context), err)
	os.Exit(1)
}

func exite(context string, input []byte, err *e.Error) {
	fmt.Printf("%s:\n\n%s\n\n", h.Bold(context), err.String(input))
	os.Exit(1)
}
