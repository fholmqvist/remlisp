package cli

import (
	"fmt"
	"os/exec"

	"github.com/fholmqvist/remlisp/compiler"
	"github.com/fholmqvist/remlisp/expander"
	"github.com/fholmqvist/remlisp/lexer"
	"github.com/fholmqvist/remlisp/parser"
)

func Run() {
	// TODO:
	printLogo()
	lexer, parser, compiler := lexer.New(), parser.New(), compiler.New()
	expander := expander.New(lexer, parser, compiler)
	stdlib := compileFile("stdlib/stdlib.rem", false, lexer, parser, expander, compiler)
	code := compileFile("input.rem", true, lexer, parser, expander, compiler)
	if err := createFile("out.js", fmt.Sprintf("%s\n\n%s", stdlib, code)); err != nil {
		exit("creating output file", err)
	}
	bb, err := exec.Command("deno", "run", "--allow-read", "out.js").Output()
	if err != nil {
		exit("deno", err)
	}
	prettyPrintResult(bb)
}
