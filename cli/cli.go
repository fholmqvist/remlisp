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
	"github.com/fholmqvist/remlisp/repl"
	"github.com/fholmqvist/remlisp/runtime"
	"github.com/fholmqvist/remlisp/transpiler"
)

func Run() {
	// TODO: Actual CLI.
	print.Logo()
	lexer, parser, transpiler := lexer.New(), parser.New(), transpiler.New()
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
			exite("compiling input.rem", input, erre)
		}
		result := fmt.Sprintf("%s\n\n// ========\n// stdlib\n// ========\n\n%s", code, stdlib)
		if err := createFile("out.js", result); err != nil {
			exit("creating output file", err)
		}
		bb, err := exec.Command("deno", "run", "--allow-read", "out.js").Output()
		if err != nil {
			exit("deno", err)
		}
		print.Result(bb)
	}
}
