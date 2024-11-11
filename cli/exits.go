package cli

import (
	"fmt"
	"os"

	e "github.com/fholmqvist/remlisp/err"
	h "github.com/fholmqvist/remlisp/highlight"
)

func exit(context string, err error) {
	fmt.Printf("%s: %s\n\n", h.Red(h.Bold("error "+context)), err)
	os.Exit(1)
}

func exite(context string, input []byte, err *e.Error) {
	fmt.Printf("%s:\n%s\n\n", h.Red(h.Bold("error "+context)), err.String(input))
	os.Exit(1)
}
