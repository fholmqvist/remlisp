package cli

import (
	"fmt"
	"os/exec"
)

func Run() {
	printLogo()
	stdlib := compileFile("stdlib/stdlib.rem", false)
	code := compileFile("input.rem", true)
	if err := createFile("out.js", fmt.Sprintf("%s\n\n%s", stdlib, code)); err != nil {
		exit("creating output file", err)
	}
	if err := exec.Command("deno", "run", "outjs").Run(); err != nil {
		exit("deno", err)
	}
}
