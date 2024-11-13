package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fholmqvist/remlisp/compiler"
	"github.com/fholmqvist/remlisp/expander"
	"github.com/fholmqvist/remlisp/lexer"
	"github.com/fholmqvist/remlisp/parser"
	"github.com/fholmqvist/remlisp/print"
	"github.com/fholmqvist/remlisp/transpiler"
)

func Run() {
	// TODO: Actual CLI.
	print.Logo()
	lexer, parser, transpiler := lexer.New(), parser.New(), transpiler.New()
	expander := expander.New(lexer, parser, transpiler)
	c := compiler.New(lexer, parser, transpiler)
	stdlibInput, stdlib, erre := c.CompileFile("stdlib/stdlib.rem", false, expander)
	if erre != nil {
		exite("compiling stdlib", stdlibInput, erre)
	}
	shouldPrint := len(os.Args) > 1 && os.Args[1] == "--debug"
	input, code, erre := c.CompileFile("input.rem", shouldPrint, expander)
	if erre != nil {
		exite("compiling input.rem", input, erre)
	}
	if err := createFile("out.js", fmt.Sprintf("%s\n\n%s", stdlib, code)); err != nil {
		exit("creating output file", err)
	}
	bb, err := exec.Command("deno", "run", "--allow-read", "out.js").Output()
	if err != nil {
		exit("deno", err)
	}
	print.Result(bb)
}
