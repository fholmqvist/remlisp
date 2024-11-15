package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fholmqvist/remlisp/compiler"
	e "github.com/fholmqvist/remlisp/err"
	"github.com/fholmqvist/remlisp/expander"
	h "github.com/fholmqvist/remlisp/highlight"
	"github.com/fholmqvist/remlisp/lexer"
	"github.com/fholmqvist/remlisp/parser"
	"github.com/fholmqvist/remlisp/print"
	"github.com/fholmqvist/remlisp/repl"
	"github.com/fholmqvist/remlisp/runtime"
	"github.com/fholmqvist/remlisp/transpiler"
)

func Run() {
	// TODO: Actual CLI.
	print.Logo()
	lexer := lexer.New()
	parser, transpiler := parser.New(lexer), transpiler.New()
	exp := expander.New(lexer, parser, transpiler)
	cmp := compiler.New(lexer, parser, transpiler)
	stdlibInput, stdlib, erre := cmp.CompileFile("stdlib/stdlib.rem", false, exp)
	if erre != nil {
		exite("compiling stdlib", stdlibInput, erre)
	}
	if len(os.Args) > 1 && os.Args[1] == "--repl" {
		rt, erre := runtime.New(cmp, exp)
		if erre != nil {
			exite("creating runtime", []byte{}, erre)
		}
		repl.Run(rt, stdlibInput)
	} else {
		shouldPrint := len(os.Args) > 1 && os.Args[1] == "--debug"
		input, code, erre := cmp.CompileFile("input.rem", shouldPrint, exp)
		if erre != nil {
			exite("reading input", input, erre)
		}
		result := fmt.Sprintf("%s\n\n// ========\n// stdlib\n// ========\n\n%s", code, stdlib)
		if err := os.WriteFile("out.js", []byte(result), os.ModePerm); err != nil {
			exit("creating output file", err)
		}
		bb, err := exec.Command("deno", "run", "--allow-read", "out.js").Output()
		if err != nil {
			exit("deno", err)
		}
		print.Result(bb)
	}
}

func exit(context string, err error) {
	fmt.Printf("%s: %s\n\n", h.Red(h.Bold("error "+context)), err)
	os.Exit(1)
}

func exite(context string, input []byte, err *e.Error) {
	fmt.Printf("%s:\n%s\n\n", h.Red(h.Bold("error "+context)), err.String(input))
	os.Exit(1)
}
