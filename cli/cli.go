package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fholmqvist/remlisp/expander"
	"github.com/fholmqvist/remlisp/lexer"
	"github.com/fholmqvist/remlisp/parser"
	compiler "github.com/fholmqvist/remlisp/transpiler"
)

func Run() {
	// TODO: Actual CLI.
	printLogo()
	lexer, parser, compiler := lexer.New(), parser.New(), compiler.New()
	expander := expander.New(lexer, parser, compiler)
	stdlib := compileFile("stdlib/stdlib.rem", false, lexer, parser, expander, compiler)
	print := len(os.Args) > 1 && os.Args[1] == "--debug"
	code := compileFile("input.rem", print, lexer, parser, expander, compiler)
	if err := createFile("out.js", fmt.Sprintf("%s\n\n%s", stdlib, code)); err != nil {
		exit("creating output file", err)
	}
	bb, err := exec.Command("deno", "run", "--allow-read", "out.js").Output()
	if err != nil {
		exit("deno", err)
	}
	prettyPrintResult(bb)
}
