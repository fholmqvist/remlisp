package cli

import (
	"fmt"
	"os/exec"
)

func Run() {
	// TODO:
	printLogo()
	stdlib := compileFile("stdlib/stdlib.rem", false)
	code := compileFile("input.rem", true)
	if err := createFile("out.js", fmt.Sprintf("%s\n\n%s", stdlib, code)); err != nil {
		exit("creating output file", err)
	}
	bb, err := exec.Command("deno", "run", "--allow-read", "out.js").Output()
	if err != nil {
		exit("deno", err)
	}
	prettyPrintResult(bb)
}
