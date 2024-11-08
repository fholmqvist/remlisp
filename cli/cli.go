package cli

import (
	"fmt"
	"os/exec"

	h "github.com/fholmqvist/remlisp/highlight"
	"github.com/fholmqvist/remlisp/pp"
)

func Run() {
	printLogo()
	stdlib := compileFile("stdlib/stdlib.rem", false)
	code := compileFile("input.rem", true)
	if err := createFile("out.js", fmt.Sprintf("%s\n\n%s", stdlib, code)); err != nil {
		exit("creating output file", err)
	}
	bb, err := exec.Command("deno", "run", "out.js").Output()
	if err != nil {
		exit("deno", err)
	}
	lisp, err := pp.FromJS(bb)
	if err != nil {
		exit("parsing output", err)
	}
	fmt.Printf("%s\n\n", h.Code(lisp))
}
