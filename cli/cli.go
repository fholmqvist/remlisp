package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/alexflint/go-arg"

	"github.com/fholmqvist/remlisp/compiler"
	e "github.com/fholmqvist/remlisp/err"
	"github.com/fholmqvist/remlisp/expander"
	h "github.com/fholmqvist/remlisp/highlight"
	"github.com/fholmqvist/remlisp/lexer"
	"github.com/fholmqvist/remlisp/parser"
	"github.com/fholmqvist/remlisp/print"
	"github.com/fholmqvist/remlisp/repl"
	"github.com/fholmqvist/remlisp/runtime"
	"github.com/fholmqvist/remlisp/stdlib"
	"github.com/fholmqvist/remlisp/transpiler"
)

func Run() {
	var settings Settings
	parg := arg.MustParse(&settings)
	lexer := lexer.New()
	parser, transpiler := parser.New(lexer), transpiler.New()
	exp := expander.New(lexer, parser, transpiler)
	cmp := compiler.New(lexer, parser, transpiler)
	if settings.REPL {
		print.Logo()
		rt, erre := runtime.New(cmp, exp)
		if erre != nil {
			exite("creating runtime", []byte{}, erre)
		}
		repl.Run(rt, stdlib.Stdlib)
	} else if settings.Path != "" {
		if settings.Debug {
			print.Logo()
		}
		std, erre := cmp.Compile(stdlib.Stdlib, exp)
		if erre != nil {
			exite("compiling stdlib", stdlib.Stdlib, erre)
		}
		input, code, erre := cmp.CompileFile(settings.Path, settings.Debug, exp)
		if erre != nil {
			exite("reading input", input, erre)
		}
		result := fmt.Sprintf("%s\n\n// ========\n// stdlib\n// ========\n\n%s", code, std)
		if err := os.WriteFile("out.js", []byte(result), os.ModePerm); err != nil {
			exit("creating output file", err)
		}
		if _, err := exec.Command("deno", "fmt", "out.js").Output(); err != nil {
			exit("deno fmt", err)
		}
		bb, err := exec.Command("deno", "run", "--allow-read", "out.js").Output()
		if err != nil {
			exit("deno", err)
		}
		print.Result(bb, settings.Debug)
	} else {
		print.Logo()
		parg.WriteUsage(os.Stdout)
		fmt.Println()
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
