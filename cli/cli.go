package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
	parg, settings, cmp, exp, rt, stdfns := setup()
	if settings.REPL {
		runRepl(cmp, exp, rt)
	} else if settings.Path != "" {
		runFile(settings, cmp, exp, stdfns)
	} else {
		showUsage(parg)
	}
}

func setup() (*arg.Parser, Settings, *compiler.Compiler, *expander.Expander, *runtime.Runtime, string) {
	var settings Settings
	parg := arg.MustParse(&settings)
	lexer := lexer.New()
	parser, transpiler := parser.New(lexer), transpiler.New()
	exp := expander.New(lexer, parser, transpiler)
	cmp := compiler.New(lexer, parser, transpiler)
	stdfns, erre := cmp.Compile(stdlib.StdFns, exp)
	if erre != nil {
		exite("compiling stdlib", stdlib.StdFns, erre)
	}
	stdmacros, erre := cmp.Compile(stdlib.StdMacros, exp)
	if erre != nil {
		exite("compiling stdlib", stdlib.StdMacros, erre)
	}
	rt, erre := runtime.New()
	if erre != nil {
		exite("creating runtime", []byte{}, erre)
	}
	rt.Send(stdfns)
	rt.Send(stdmacros)
	return parg, settings, cmp, exp, rt, stdfns
}

func runRepl(cmp *compiler.Compiler, exp *expander.Expander, rt *runtime.Runtime) {
	print.Logo()
	repl.Run(cmp, exp, rt)
}

func runFile(settings Settings, cmp *compiler.Compiler, exp *expander.Expander, stdfns string) {
	if settings.Debug {
		print.Logo()
	}
	input, code, erre := cmp.CompileFile(settings.Path, settings.Debug, exp)
	if erre != nil {
		exite("reading input", input, erre)
	}
	result := fmt.Sprintf("%s\n\n// ========\n// stdlib\n// ========\n\n%s", code, stdfns)
	outfile := "out.js"
	if settings.Out != "" {
		outfile = settings.Out
		if !strings.HasSuffix(outfile, ".js") {
			outfile += ".js"
		}
	}
	if err := os.WriteFile(outfile, []byte(result), os.ModePerm); err != nil {
		exit("creating output file", err)
	}
	if _, err := exec.Command("deno", "fmt", outfile).Output(); err != nil {
		exit("deno fmt", err)
	}
	if settings.Run {
		bb, err := exec.Command("deno", "run", "--allow-read", outfile).Output()
		if err != nil {
			exit("deno", err)
		}
		print.Result(bb, settings.Debug)
	}
}

func showUsage(parg *arg.Parser) {
	print.Logo()
	parg.WriteUsage(os.Stdout)
	fmt.Println()
}

func exit(context string, err error) {
	fmt.Printf("%s: %s\n\n", h.Red(h.Bold("error "+context)), err)
	os.Exit(1)
}

func exite(context string, input []byte, err *e.Error) {
	fmt.Printf("%s:\n%s\n\n", h.Red(h.Bold("error "+context)), err.String(input))
	os.Exit(1)
}
